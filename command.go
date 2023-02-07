// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"context"
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

	// commandName is the base name of the executable.
	commandName = filepath.Base(executable)

	// commandLine contains the full path to the command and each argument.
	commandLine = append([]string{executable}, os.Args[1:]...)

	// module identifies the module's package path.
	module string

	// vmmp: version.major.minor-timestamp-commithash
	vmmp string

	// buildDate sets the build date for the command.
	buildDate = func() string {
		info, _ := os.Stat(executable)
		return info.ModTime().UTC().Format("2006-01-02 15:04:05 UTC")
	}()
)

// Main drives the show.
func Main(main func(context.Context)) {
	module, vmmp = build()

	if !parse(os.Args[1:]) {
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
		main(ctx) // if service exits, proceed with cleanup
		stop()    // inform service routines of exit for cleanup
	}()

	// run osEnvironment on main thread for the native host application environment setup (e.g. MacOS main run loop)
	// osEnvironment(ctx)

	<-ctx.Done()
	<-time.After(time.Millisecond * 1000) // wait a moment for contexts to cleanup and exit
}

// build gathers the module and version information for this build.
func build() (string, string) {
	_, n, _, _ := runtime.Caller(2)
	dir := filepath.Dir(n)
	mod := Module(dir)
	_, vers, ok := strings.Cut(mod.Dir, "@")
	if !ok {
		cmd := exec.Command("git", "show", "-s", "--format=%cI %H")
		cmd.Dir = mod.Dir
		out, err := cmd.Output()
		if err == nil {
			tm, h, _ := strings.Cut(string(out), " ")
			t, err := time.Parse(time.RFC3339, tm)
			if err != nil {
				panic(fmt.Errorf("time parse failed %s %v", out, err))
			}
			vers = t.UTC().Format("v0.0.0-20060102150405-") + h[:12]
		} else {
			vers = time.Now().UTC().Format("v0.0.0-20060102150405-000000000000")
		}
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
