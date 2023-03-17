// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type (
	// Err custom logging error type
	Err struct {
		user   string
		file   string
		line   int
		detail string
		err    error
	}
)

const (
	levelTrace = iota - 2
	levelDebug
	levelInfo // default
	levelWarn
	levelError
)

var (
	// logLeveel maps parsed log levels to level index.
	logLevel = func() int {
		switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
		case "TRACE":
			return levelTrace
		case "DEBUG":
			return levelDebug
		case "WARN":
			return levelWarn
		case "ERROR":
			return levelError
		}
		return levelInfo
	}()
)

// Error method to comply with error interface
func (err *Err) Error() string {
	return fmt.Sprintf("%s [%s %s %s] [%s] [%s:%d] %s%s",
		os.Args,
		executable,
		vmmp,
		buildDate,
		err.user,
		err.file,
		err.line,
		err.detail,
		err.err.Error(),
	)
}

// Unwrap method to comply with error interface
func (err *Err) Unwrap() error {
	return err.err
}

// Error formats an error with function name and error message, with code location
// details for initial error, preserving the initial logged error for percolation.
func Error(name string, err error) *Err {
	return logMessage(2, name, err)
}

// Unsupported reports that a specific OS does not support a function
func Unsupported() error {
	return Error("Unsupported", fmt.Errorf(runtime.GOOS))
}

// logWrite writes a log message to the log destination.
func logWrite(level string, name string, err error) {
	if err := logMessage(3, name, err); err != nil {
		log.Printf("%s %-5s %s", time.Now().Format(TimeFormat), level, err.Error())
	}
}

// logMessage formats a log message with where, what, how, which, who, why of an error.
func logMessage(depth int, name string, err error) *Err {
	if err == nil {
		return nil
	}
	e := &Err{}
	if errors.As(err, &e) {
		return e // percolate original Err
	}

	_, file, line, _ := runtime.Caller(depth)

	var detail string
	if name != "" {
		detail += name + ": "
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		detail += fmt.Sprintf("errno %d: ", errno)
	}

	return &Err{
		user:   Username(os.Getuid()),
		file:   file,
		line:   line,
		detail: detail,
		err:    err,
	}
}

// LogTrace log trace message.
func LogTrace(name string, err error) {
	if logLevel <= levelTrace {
		logWrite("TRACE", name, err)
	}
}

// LogDebug log debug message.
func LogDebug(name string, err error) {
	if logLevel <= levelDebug {
		logWrite("DEBUG", name, err)
	}
}

// LogInfo log info message (default logging level).
func LogInfo(name string, err error) {
	if logLevel <= levelInfo {
		logWrite("INFO", name, err)
	}
}

// LogWarn log warning message, setting exit code to WARN.
func LogWarn(name string, err error) {
	if logLevel <= levelWarn {
		logWrite("WARN", name, err)
	}
}

// LogError log error message, setting exit code to ERROR.
func LogError(name string, err error) {
	if logLevel <= levelError {
		logWrite("ERROR", name, err)
	}
}
