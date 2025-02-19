// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"bytes"
	"cmp"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type (
	flags struct {
		flag.FlagSet
		version              bool
		cpuprofile           bool
		memprofile           bool
		CommandDescription   string
		ArgumentDescriptions [][2]string
		argsMax              int
	}
)

var (
	// Flags defines and initializes the command line flags
	Flags = flags{
		FlagSet:              flag.FlagSet{},
		version:              false,
		cpuprofile:           false,
		memprofile:           false,
		CommandDescription:   "",
		ArgumentDescriptions: [][2]string{},
		argsMax:              0,
	}

	// flagSyntax is a map of flags' names to their command line syntax
	flagSyntax = map[string]string{}

	// logBuf captures error output from Go flag parser
	logBuf = bytes.Buffer{}
)

// Var maps a flag field to its name and description, and adds a brief description
func (f *flags) Var(field any, name, syntax, detail string) {
	flagSyntax[name] = syntax
	switch field := field.(type) {
	case *int:
		f.IntVar(field, name, *field, detail)
	case *uint:
		f.UintVar(field, name, *field, detail)
	case *int64:
		f.Int64Var(field, name, *field, detail)
	case *uint64:
		f.Uint64Var(field, name, *field, detail)
	case *float64:
		f.Float64Var(field, name, *field, detail)
	case *string:
		f.StringVar(field, name, *field, detail)
	case *bool:
		f.BoolVar(field, name, *field, detail)
	case *time.Duration:
		f.DurationVar(field, name, *field, detail)
	default:
		f.FlagSet.Var(field.(flag.Value), name, detail)
	}
}

// init initializes the gocore command line flags.
func init() {
	log.SetFlags(0)

	Flags.Var(
		&Flags.version,
		"version",
		"[-version]",
		"Print version information and exit",
	)

	Flags.Var(
		&Flags.cpuprofile,
		"cpuprofile",
		"[-cpuprofile]",
		"Capture a CPU performance profile for this invocation",
	)

	Flags.Var(
		&Flags.memprofile,
		"memprofile",
		"[-memprofile]",
		"Capture a memory usage profile for this invocation",
	)

	Flags.SetOutput(&logBuf) // capture FlagSet.Parse messages
	Flags.Usage = usage
}

// parse inspects the command line.
func parse(args []string) error {
	if err := Flags.Parse(args); err != nil {
		return Error("argument parser", err)
	}

	if Flags.NArg() > Flags.argsMax { // too many arguments?
		args := strings.Join(Flags.Args()[Flags.NArg()-Flags.argsMax-1:], " ")
		return Error("argument parser", fmt.Errorf("%s", args))
	}

	return nil
}

// usage formats the flags Usage message for gomon.
func usage() {
	if !IsTerminal(os.Stderr) && logBuf.Len() > 0 { // if called by go's flag package parser, may have error text
		Error("terminal", errors.New(strings.TrimSpace(logBuf.String()))).Err() // in that case report it
		return
	}

	logBuf.WriteString("NAME:\n  " + filepath.Base(Executable))
	logBuf.WriteString("\n\nDESCRIPTION:\n  " + Flags.CommandDescription)

	var flags []string
	for _, flag := range Ordered(flagSyntax, cmp.Compare) {
		flags = append(flags, flag)
	}
	logBuf.WriteString("\n\nUSAGE:\n  " + filepath.Base(Executable) + " [-help] " + strings.Join(flags, " "))

	if len(Flags.ArgumentDescriptions) > 0 {
		for _, args := range Flags.ArgumentDescriptions {
			logBuf.WriteString(" [" + args[0] + "]")
		}
	}
	logBuf.WriteString(`

VERSION:
  ` + module + " " + Version + `

OPTIONS:
  -help
	Print the help and exit
`)
	Flags.PrintDefaults()

	if len(Flags.ArgumentDescriptions) > 0 {
		logBuf.WriteString("\nARGUMENTS:\n")
		for _, args := range Flags.ArgumentDescriptions {
			logBuf.WriteString("  " + args[0] + "\n\t" + args[1] + "\n")
		}
	}
	logBuf.WriteString("\nCopyright © 2023 The Gomon Project.\n")
	fmt.Fprint(os.Stderr, logBuf.String())
}

// Regexp is a command line flag type.
type Regexp struct {
	*regexp.Regexp
}

// Set is a flag.Value interface method to enable Regexp as a command line flag.
func (r *Regexp) Set(pattern string) (err error) {
	r.Regexp, err = regexp.Compile(pattern)
	return
}

// String is a flag.Value interface method to enable Regexp as a command line flag.
func (r *Regexp) String() string {
	if r.Regexp == nil {
		return ""
	}
	return r.Regexp.String()
}
