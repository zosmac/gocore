// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type (
	// LogLevel indexes the literal log levels.
	LogLevel int

	// LogMessage is custom logging error type.
	LogMessage struct {
		Source string
		E      error
		File   string
		Line   int
		Detail map[string]string
	}
)

const (
	// enum of logging levels corresponding to log levels.
	LevelTrace LogLevel = iota - 2
	LevelDebug
	LevelInfo // default
	LevelWarn
	LevelError
	LevelFatal
)

var (
	// logLevels map logging level enums to log levels.
	logLevels = map[LogLevel]string{
		LevelTrace: "TRACE",
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
		LevelFatal: "FATAL",
	}

	// loggingLevel maps requested LOG_LEVEL to index.
	LoggingLevel = func() LogLevel {
		switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
		case "TRACE":
			return LevelTrace
		case "DEBUG":
			return LevelDebug
		case "WARN":
			return LevelWarn
		case "ERROR":
			return LevelError
		case "FATAL":
			return LevelFatal
		}
		return LevelInfo
	}()

	// Log is the default log message formatter and writer.
	Log = func(msg LogMessage, level LogLevel) {
		if level >= LoggingLevel {
			if msg.E == nil && level > LevelInfo {
				level = LevelInfo
			}
			log.Printf("%s %-5s %s", time.Now().Format(RFC3339Milli), logLevels[level], msg.Error())
		}
	}
)

// Unsupported reports that a specific OS does not support a function
func Unsupported() error {
	return Error("Unsupported", errors.New(runtime.GOOS))
}

// Error method to comply with error interface.
func (msg LogMessage) Error() string {
	var detail string
	var keys []string
	for key := range msg.Detail {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		val := msg.Detail[key]
		if val == "" {
			continue
		}
		if strings.ContainsAny(val, "\t\n\v\f\r ") { // quote value with embedded whitespace
			val = "\"" + val + "\""
		}
		detail += " " + key + "=" + val
	}
	var e string
	if msg.E != nil {
		e = "err=\"" + msg.E.Error() + "\" "
	}
	return fmt.Sprintf("[%s:%d] %ssource=%q%s",
		msg.File,
		msg.Line,
		e,
		msg.Source,
		detail,
	)
}

// Unwrap method to comply with error interface.
func (msg LogMessage) Unwrap() error {
	return msg.E
}

// Error records the function source, error message, code location, and any
// details of initial error, preserving the initial error for percolation.
func Error(source string, err error, details ...map[string]string) LogMessage {
	e := LogMessage{}
	if errors.As(err, &e) {
		return e // percolate original Err
	}

	_, file, line, _ := runtime.Caller(1)
	file = filepath.Join(filepath.Base(filepath.Dir(file)), filepath.Base(file))

	detail := map[string]string{}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		detail["errno"] = strconv.Itoa(int(errno))
	}

	// if for some reason, multiple detail maps, merge into first one.
	for _, det := range details {
		for key, val := range det {
			detail[key] = val
		}
	}

	return LogMessage{
		Source: source,
		E:      err,
		File:   file,
		Line:   line,
		Detail: detail,
	}
}

// Trace log trace message.
func (msg LogMessage) Trace() {
	Log(msg, LevelTrace)
}

// Debug log debug message.
func (msg LogMessage) Debug() {
	Log(msg, LevelDebug)
}

// Info log info message (default logging level).
func (msg LogMessage) Info() {
	Log(msg, LevelInfo)
}

// Warn log warning message.
func (msg LogMessage) Warn() {
	Log(msg, LevelWarn)
}

// Error log error message.
func (msg LogMessage) Err() {
	Log(msg, LevelError)
}
