// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

// profile turns on CPU performance or Memory usage profiling of command.
// Profiling can also be enabled via the /debug/pprof endpoint.
func profile(ctx context.Context) {
	if Flags.cpuprofile {
		if f, err := os.CreateTemp("", "pprof_"); err != nil {
			Error("cpuprofile", err).Err()
		} else {
			go func() {
				pprof.StartCPUProfile(f)
				<-ctx.Done()
				pprof.StopCPUProfile()
				cmd, _ := os.Executable()
				fmt.Fprintf(os.Stderr,
					"CPU profile written to %[1]q.\nUse the following command to evaluate:\n"+
						"\033[1;31mgo tool pprof -web %[2]s %[1]s\033[0m\n",
					f.Name(),
					cmd,
				)
				f.Close()
			}()
		}
	}

	if Flags.memprofile {
		if f, err := os.CreateTemp(".", "mprof_"); err != nil {
			Error("memprofile", err).Err()
		} else {
			go func() {
				<-ctx.Done()
				runtime.GC()
				pprof.WriteHeapProfile(f)
				cmd, _ := os.Executable()
				fmt.Fprintf(os.Stderr,
					"Memory profile written to %[1]q.\nUse the following command to evaluate:\n"+
						"\033[1;31mgo tool pprof -web %[2]s %[1]s\033[0m\n",
					f.Name(),
					cmd,
				)
				f.Close()
			}()
		}
	}
}
