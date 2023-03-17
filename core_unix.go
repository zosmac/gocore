// Copyright © 2021-2023 The Gomon Project.

//go:build !windows

package gocore

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	// euid gets the executable file's owner id.
	euid = os.Geteuid()
)

// signalContext returns context for detecting interrupt signal.
func signalContext() (context.Context, context.CancelFunc) {
	// ignore these signals to enable to continue running
	signal.Ignore(syscall.SIGWINCH, syscall.SIGHUP, syscall.SIGTTIN, syscall.SIGTTOU)
	return signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
}

// seteuid current process to file owner.
func Seteuid() {
	err := syscall.Seteuid(euid)
	LogInfo("Seteuid", fmt.Errorf("uid: %d, euid: %d err: %v",
		os.Getuid(),
		os.Geteuid(),
		err,
	))
}

// setuid current process to user.
func Setuid() {
	err := syscall.Seteuid(os.Getuid())
	LogInfo("Setuid", fmt.Errorf("uid: %d, euid: %d err: %v",
		os.Getuid(),
		os.Geteuid(),
		err,
	))
}
