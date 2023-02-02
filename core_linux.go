// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"bufio"
	"context"
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
			LogError(Error("/proc/stat open", err))
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
					LogError(Error("/proc/stat btime", err))
					return time.Time{}
				}
				return time.Unix(int64(sec), 0)
			}
		}

		LogError(Error("/proc/stat btime", sc.Err()))
		return time.Time{}
	}()
)

// GoStringN interprets a null terminated C char array as a GO string.
func GoStringN[S int8 | byte, L int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64](p *S, l L) string {
	s := make([]byte, l)
	var i L
	for ; i < l; i++ {
		if *p == 0 {
			break
		}
		s[i] = byte(*p)
		p = (*S)(unsafe.Add(unsafe.Pointer(p), 1))
	}
	return string(s[:i])
}

// FdPath gets the path for an open file descriptor.
func FdPath(fd int) (string, error) {
	pid := os.Getpid()
	return os.Readlink(filepath.Join("/proc", strconv.Itoa(pid), "fd", strconv.Itoa(fd)))
}

// MountMap builds a map of mount points to file systems.
func MountMap() (map[string]string, error) {
	f, err := os.Open("/etc/mtab")
	if err != nil {
		return nil, err
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
