// Copyright © 2021-2023 The Gomon Project.

package gocore

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
)

var (
	// Host identifies the local host.
	Host, _ = os.Hostname()

	// executable identifies the full command path.
	executable, _ = os.Executable()

	// module identifies the module's package path.
	module string

	// vmmp: version.major.minor-timestamp-commithash
	vmmp string

	// buildDate sets the build date for the command.
	buildDate = func() string {
		info, _ := os.Stat(executable)
		return info.ModTime().UTC().Format("2006-01-02T15:04:05Z")
	}()
)

// Main drives the show.
func Main(main func(context.Context) error) {
	module, vmmp = build()

	if err := parse(os.Args[1:]); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			LogError("", err)
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
			LogError("", err)
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
		executable, module, vmmp, buildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
