// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"unsafe"
)

type (
	// ValidValue defines list of values that are valid for a type safe string.
	ValidValue[T ~string] map[T]int
)

const (
	// TimeFormat used for formatting timestamps.
	TimeFormat = "2006-01-02T15:04:05.000Z07:00"
)

var (
	// HostEndian enables byte order conversion between local and network integers.
	HostEndian = func() binary.ByteOrder {
		n := uint16(1)
		a := (*[2]byte)(unsafe.Pointer(&n))[:]
		b := []byte{0, 1}
		if bytes.Equal(a, b) {
			return binary.BigEndian
		}
		return binary.LittleEndian
	}()
)

// Define initializes a ValidValue type with its valid values.
func (vv ValidValue[T]) Define(values ...T) ValidValue[T] {
	vv = map[T]int{}
	for i, v := range values {
		vv[v] = i
	}
	return vv
}

// ValidValues returns an ordered list of valid values for the type.
func (vv ValidValue[T]) ValidValues() []string {
	ss := make([]string, len(vv))
	for v, i := range vv {
		ss[i] = string(v)
	}
	return ss
}

// IsValid returns whether a value is valid.
func (vv ValidValue[T]) IsValid(v T) bool {
	_, ok := vv[v]
	return ok
}

// Index returns the position of a value in the valid value list.
func (vv ValidValue[T]) Index(v T) int {
	return vv[v]
}

// ChDir is a convenience function for changing the current directory and reporting its canonical path.
// If changing the directory fails, ChDir returns the error and canonical path of the current directory.
func ChDir(dir string) (string, error) {
	var err error
	if dir, err = filepath.Abs(dir); err == nil {
		if dir, err = filepath.EvalSymlinks(dir); err == nil {
			if err = os.Chdir(dir); err == nil {
				return dir, nil
			}
		}
	}
	dir, _ = os.Getwd()
	dir, _ = filepath.EvalSymlinks(dir)
	return dir, err
}

// IsTerminal reports if a file handle is connected to the terminal.
func IsTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	terminal := os.ModeDevice | os.ModeCharDevice
	return info.Mode()&terminal == terminal
}

// StartCommand starts a command and returns a scanner for reading stdout.
func StartCommand(ctx context.Context, cmdline []string) (*bufio.Scanner, error) {
	cmd := exec.CommandContext(ctx, cmdline[0], cmdline[1:]...)

	cmd.ExtraFiles = extraFiles()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, Error("StdoutPipe()", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, Error("Start()", err)
	}

	LogInfo(fmt.Errorf(
		"Start() command=%q pid=%d",
		cmd.String(),
		cmd.Process.Pid,
	))

	go wait(cmd)

	return bufio.NewScanner(stdout), nil
}

// wait for a started command to complete and report its exit status.
func wait(cmd *exec.Cmd) {
	err := cmd.Wait()
	state := cmd.ProcessState
	var stderr string
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			stderr = string(err.Stderr)
		}
	}
	// log.DefaultLogger.Info(
	LogInfo(fmt.Errorf(
		"Wait() command=%q pid=%d err=%v rc=%d systime=%v, usrtime=%v stderr=%s", // \nsys=%#v usage=%#v",
		cmd.String(),
		cmd.Process.Pid,
		err,
		state.ExitCode(),
		state.SystemTime(),
		state.UserTime(),
		stderr,
		// state.Sys(),
		// state.SysUsage(),
	))
}
