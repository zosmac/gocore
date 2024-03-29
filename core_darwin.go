// Copyright © 2021-2023 The Gomon Project.

package gocore

/*
#cgo CFLAGS: -x objective-c -std=gnu11 -fobjc-arc
#cgo LDFLAGS: -framework CoreFoundation
#import <CoreFoundation/CoreFoundation.h>
#include <libproc.h>
#include <sys/sysctl.h>

// createCFString required to avoid go vet warning "possible misuse of unsafe.Pointer".
static void *
createCFString(char *s) {
	return (void*)CFStringCreateWithCString(nil, s, kCFStringEncodingUTF8);
}
*/
import "C"

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// CGO interprets C. types in different packages as being different types. The following gocore package aliases for
// CoreFoundation (C.CF...) types externalize these locally defined types as gocore package types. Casting C.CF...Ref
// arguments to gocore.CF...Ref enables callers from other packages by using the gocore package type name.
type (
	// CFTypeRef creates gocore package alias for type
	CFTypeRef = C.CFTypeRef
	// CFStringRef creates gocore package alias for type
	CFStringRef = C.CFStringRef
	// CFNumberRef creates gocore package alias for type
	CFNumberRef = C.CFNumberRef
	// CFBooleanRef creates gocore package alias for type
	CFBooleanRef = C.CFBooleanRef
	// CFArrayRef creates gocore package alias for type
	CFArrayRef = C.CFArrayRef
	// CFDictionaryRef creates gocore package alias for type
	CFDictionaryRef = C.CFDictionaryRef
)

var (
	// Boottime retrieves the system boot time.
	Boottime = func() time.Time {
		var timespec C.struct_timespec
		size := C.size_t(C.sizeof_struct_timespec)
		if rv, err := C.sysctl(
			&[]C.int{C.CTL_KERN, C.KERN_BOOTTIME}[0],
			2,
			unsafe.Pointer(&timespec),
			&size,
			unsafe.Pointer(nil),
			0,
		); rv != 0 {
			Error("sysctl kern.boottime", err).Err()
			return time.Time{}
		}

		return time.Unix(int64(timespec.tv_sec), int64(timespec.tv_nsec))
	}()
)

// FdPath gets the path for an open file descriptor.
func FdPath(fd int) (string, error) {
	var fdi C.struct_vnode_fdinfowithpath
	if n, err := C.proc_pidfdinfo(
		C.int(os.Getpid()),
		C.int(fd),
		C.PROC_PIDFDVNODEPATHINFO,
		unsafe.Pointer(&fdi),
		C.PROC_PIDFDVNODEPATHINFO_SIZE,
	); n <= 0 {
		return "", Error("proc_pidfdinfo PROC_PIDFDVNODEPATHINFO", err)
	}
	return C.GoString(&fdi.pvip.vip_path[0]), nil
}

// MountMap builds a map of mount points to file systems.
func MountMap() (map[string]string, error) {
	n, err := syscall.Getfsstat(nil, C.MNT_NOWAIT)
	if err != nil {
		return nil, Error("getfsstat", err)
	}
	list := make([]syscall.Statfs_t, n)
	if _, err = syscall.Getfsstat(list, C.MNT_NOWAIT); err != nil {
		return nil, Error("getfsstat", err)
	}

	m := map[string]string{"/": ""} // have "/" at a minimum
	for _, f := range list[0:n] {
		if f.Blocks == 0 {
			continue
		}
		m[C.GoString((*C.char)(&f.Mntonname[0]))] =
			C.GoString((*C.char)(&f.Mntfromname[0]))
	}
	return m, nil
}

// CreateCFString copies a Go string as a Core Foundation CFString. Requires CFRelease be called when done.
func CreateCFString(s string) unsafe.Pointer {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return unsafe.Pointer(C.createCFString(cs))
}

// GetCFString gets a Go string from a CFString
func GetCFString(p CFStringRef) string {
	if p == 0 {
		return ""
	}

	if s := C.CFStringGetCStringPtr(p, C.kCFStringEncodingUTF8); s != nil {
		return C.GoString(s)
	}

	var buf [1024]C.char
	C.CFStringGetCString(
		p,
		&buf[0],
		C.CFIndex(len(buf)),
		C.kCFStringEncodingUTF8,
	)
	return C.GoString(&buf[0])
}

// GetCFNumber gets a Go numeric type from a CFNumber
func GetCFNumber(n CFNumberRef) any {
	var i int64
	var f float64
	t := C.CFNumberType(C.kCFNumberSInt64Type)
	p := unsafe.Pointer(&i)
	v := any(&i)
	if C.CFNumberIsFloatType(n) == C.true {
		t = C.kCFNumberFloat64Type
		p = unsafe.Pointer(&f)
		v = any(&f)
	}
	C.CFNumberGetValue(n, t, p)
	if _, ok := v.(*int64); ok {
		return i
	}
	return f
}

// GetCFBoolean gets a Go bool from a CFBoolean
func GetCFBoolean(b CFBooleanRef) bool {
	return C.CFBooleanGetValue(b) != 0
}

// GetCFArray gets a Go slice from a CFArray
func GetCFArray(a CFArrayRef) []any {
	c := C.CFArrayGetCount(a)
	s := make([]any, c)
	vs := make([]unsafe.Pointer, c)
	C.CFArrayGetValues(a, C.CFRange{length: c, location: 0}, &vs[0])

	for i, v := range vs {
		s[i] = GetCFValue(v)
	}

	return s
}

// GetCFDictionary gets a Go map from a CFDictionary
func GetCFDictionary(d CFDictionaryRef) map[string]any {
	if d == 0 {
		return nil
	}
	c := int(C.CFDictionaryGetCount(d))
	m := make(map[string]any, c)
	ks := make([]unsafe.Pointer, c)
	vs := make([]unsafe.Pointer, c)
	C.CFDictionaryGetKeysAndValues(d, &ks[0], &vs[0])

	for i, k := range ks {
		if C.CFGetTypeID(C.CFTypeRef(k)) != C.CFStringGetTypeID() {
			continue
		}
		m[GetCFString(CFStringRef(k))] = GetCFValue(vs[i])
	}

	return m
}

func GetCFValue(v unsafe.Pointer) any {
	switch id := C.CFGetTypeID(C.CFTypeRef(v)); id {
	case C.CFStringGetTypeID():
		return GetCFString(CFStringRef(v))
	case C.CFNumberGetTypeID():
		return GetCFNumber(CFNumberRef(v))
	case C.CFBooleanGetTypeID():
		return GetCFBoolean(CFBooleanRef(v))
	case C.CFDictionaryGetTypeID():
		return GetCFDictionary(CFDictionaryRef(v))
	case C.CFArrayGetTypeID():
		return GetCFArray(CFArrayRef(v))
	default:
		d := C.CFCopyDescription(C.CFTypeRef(v))
		t := GetCFString(d)
		C.CFRelease(C.CFTypeRef(d))
		return fmt.Sprintf("Unrecognized Type is %d: %s\n", id, t)
	}
}

// darkmode is called from viewDidChangeEffectiveAppearance to report changes to the system appearance.
// It must be defined separately from the declaration in core_darwin.go to prevent duplicate symbol link error.
// From the CGO documentation (https://golang.google.cn/cmd/cgo#hdr-C_references_to_Go):
//
//	Using //export in a file places a restriction on the preamble: since it is copied into two different
//	C output files, it must not contain any definitions, only declarations. If a file contains both definitions
//	and declarations, then the two output files will produce duplicate symbols and the linker will fail. To avoid
//	this, definitions must be placed in preambles in other files, or in C source files.
//

//export darkmode
func darkmode(dark C.bool) {
	DarkAppearance = bool(dark)
}

// extraFiles called by Spawn to nil fds beyond 2 (stderr) so that they are closed on exec.
func extraFiles() []*os.File {
	// ensure that no open descriptors propagate to child
	if n := C.proc_pidinfo(
		C.int(os.Getpid()),
		C.PROC_PIDLISTFDS,
		0,
		nil,
		0,
	); n >= 3*C.PROC_PIDLISTFD_SIZE {
		return make([]*os.File, (n/C.PROC_PIDLISTFD_SIZE)-3) // close gomon files in child
	}

	return nil
}
