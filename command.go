// Copyright © 2021-2023 The Gomon Project.

package gocore

import "C"

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

type (
	// Utsname contains Go format system uname
	Utsname struct {
		Sysname  string
		Nodename string
		Release  string
		Version  string
		Machine  string
	}
)

var (
	// Host identifies the local host.
	Host, _ = os.Hostname()

	// Platform identifies the local OS.
	Platform = runtime.GOOS + "_" + runtime.GOARCH

	Uname = func() Utsname {
		var utsname unix.Utsname
		if err := unix.Uname(&utsname); err == nil {
			return Utsname{
				Sysname:  C.GoString((*C.char)(unsafe.Pointer(&utsname.Sysname[0]))),
				Nodename: C.GoString((*C.char)(unsafe.Pointer(&utsname.Nodename[0]))),
				Release:  C.GoString((*C.char)(unsafe.Pointer(&utsname.Release[0]))),
				Version:  C.GoString((*C.char)(unsafe.Pointer(&utsname.Version[0]))),
				Machine:  C.GoString((*C.char)(unsafe.Pointer(&utsname.Machine[0]))),
			}
		}
		return Utsname{}
	}()

	// executable identifies the full command path.
	Executable, _ = os.Executable()

	// module identifies the module's package path.
	module string

	// Version of module: version.major.minor-timestamp-commithash
	Version string

	// buildDate sets the build date for the command.
	buildDate = func() string {
		info, _ := os.Stat(Executable)
		return info.ModTime().UTC().Format("2006-01-02T15:04:05Z")
	}()
)

// Main drives the show.
func Main(main func(context.Context) error) {
	module, Version = build()

	if err := parse(os.Args[1:]); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			Error("", err).Err()
		}
		return
	}

	if Flags.version {
		version()
		return
	}

	// ctx, cncl := context.WithCancel(context.Background())
	ctx, stop := signalContext()

	// set up profiling if requested
	profile(ctx)

	go func() {
		if err := main(ctx); err != nil {
			Error("exit maini", err).Err()
		}
		stop() // on exit, inform service routines to cleanup
	}()

	// run osEnvironment on main thread for the native host application environment setup (e.g. MacOS main run loop)
	// osEnvironment(ctx)

	<-ctx.Done()
	<-time.After(time.Millisecond * 1000) // wait a moment for contexts to cleanup and exit
}

// build gathers the module and version information for this build.
func build() (string, string) {
	_, file, _, _ := runtime.Caller(2)
	mod := Module(filepath.Dir(file))
	_, vers, ok := strings.Cut(mod.Dir, "@")
	if !ok {
		// get git repo time and hash
		cmd := exec.Command("git", "show", "-s", "--format=%cI %H")
		cmd.Dir = mod.Dir
		out, _ := cmd.Output()
		tm, hash, _ := strings.Cut(string(out), " ")
		t, _ := time.Parse(time.RFC3339, tm)
		hash = strings.TrimSpace(hash) + strings.Repeat("0", 12)
		vers = "v0.0.0-" + t.UTC().Format("20060102150405-") + hash[:12]
	}

	return mod.Path, vers
}

// version returns the command's version information.
func version() {
	fmt.Fprintf(os.Stderr,
		`Command    - %s
Module     - %s
Version    - %s
Build Date - %s
Compiler   - %s %s_%s
Copyright © 2021-2023 The Gomon Project.
`,
		Executable, module, Version, buildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
