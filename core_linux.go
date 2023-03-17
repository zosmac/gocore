// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var (
	// Boottime gets the system boot time.
	Boottime = func() time.Time {
		f, err := os.Open("/proc/stat")
		if err != nil {
			LogError("/proc/stat open", err)
			return time.Time{}
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			l := sc.Text()
			k, v, _ := strings.Cut(l, " ")
			switch k {
			case "btime":
				sec, err := strconv.Atoi(v)
				if err != nil {
					LogError("/proc/stat btime", err)
					return time.Time{}
				}
				return time.Unix(int64(sec), 0)
			}
		}

		LogError("/proc/stat btime", sc.Err())
		return time.Time{}
	}()
)

// GoStringN interprets a null terminated C char array as a GO string.
func GoString[C int8 | byte](char *C) string {
	buf := []byte{}
	for *char != 0 {
		buf = append(buf, byte(*char))
		char = (*C)(unsafe.Add(unsafe.Pointer(char), 1))
	}
	return string(buf)
}

// GoStringN interprets a length specified and possibly null terminated C char array as a GO string.
func GoStringN[
	C int8 | byte,
	L int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64,
](char *C, l L) string {
	if l == 0 {
		return GoString(char)
	}
	buf := make([]byte, l)
	cs := unsafe.Slice(char, l)
	var c C
	var i int
	for i, c = range cs {
		if c == 0 {
			break
		}
		buf[i] = byte(c)
	}
	return string(buf[:i])
}

// FdPath gets the path for an open file descriptor.
func FdPath(fd int) (string, error) {
	return os.Readlink(filepath.Join("/proc", "self", "fd", strconv.Itoa(fd)))
}

// MountMap builds a map of mount points to file systems.
func MountMap() (map[string]string, error) {
	f, err := os.Open("/etc/mtab")
	if err != nil {
		return nil, Error("Open /etc/mtab", err)
	}
	defer f.Close()
	m := map[string]string{"/": ""} // have "/" at a minimum
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		m[f[1]] = f[0]
	}
	return m, nil
}

// Measures reads a /proc filesystem file and produces a map of name:value pairs.
func Measures(filename string) (map[string]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if k, v, ok := strings.Cut(sc.Text(), ":"); ok {
			v := strings.Fields(v)
			if len(v) > 0 {
				m[k] = v[0]
			}
		}
	}

	return m, nil
}

// extraFiles called by Spawn to nil fds beyond 2 (stderr) so that they are closed on exec.
func extraFiles() []*os.File {
	dirname := filepath.Join("/proc", "self", "fd")
	if dir, err := os.Open(dirname); err == nil {
		fds, err := dir.Readdirnames(0)
		dir.Close()
		if err == nil {
			maxFd := -1
			for _, fd := range fds {
				if n, err := strconv.Atoi(fd); err == nil && n > maxFd {
					maxFd = n
				}
			}
			// ensure that no open descriptors propagate to child
			if maxFd >= 3 {
				return make([]*os.File, maxFd-3) // close gomon files in child
			}
		}
	}

	return nil
}
