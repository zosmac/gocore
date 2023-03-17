// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"errors"
	"net"
	"net/netip"
	"os/user"
	"runtime"
	"strconv"
	"time"

	"golang.org/x/tools/go/packages"
)

type (
	// uname is key to cached user name.
	uname int

	// gname is key to cached group name.
	gname int

	// hname is key to cached host name.
	hname string

	// moddir is key to cached go module information.
	moddir string

	// modval is cached value of module information.
	modval struct {
		Dir  string
		Path string
		Pkg  string
	}
)

var (
	// DarkAppearance indicates whether system appearance is "dark" or "light"
	DarkAppearance bool

	// unames is the cache of user names.
	unames = newCache(username)

	// gnames is the cache of group names.
	gnames = newCache(groupname)

	// hnames is the cache of host names.
	hnames = newCache(hostname)

	// mnames is the cache of go module information.
	mnames = newCache(modinfo)
)

// MsToTime converts Unix era milliseconds to Go time.Time.
func MsToTime(ms uint64) time.Time {
	var s, n int64
	if runtime.GOOS == "windows" {
		t := ms * 1e6                  // ns since 1/1/1601 overflows int64, use uint64
		s = int64(t/1e9 - 11644473600) // 1/1/1970 - 1/1/1601 offset in seconds
		n = int64(t % 1e9)
	} else {
		s = int64(ms / 1e3)       // truncate ms to s
		n = int64(ms % 1e3 * 1e6) // compute ns remainder
	}

	return time.Unix(s, n)
}

// Username retrieves and caches user name for uid.
func Username(uid int) string {
	value, _ := unames.lookup(uname(uid))
	return value
}

// Groupname retrieves and caches group name for gid.
func Groupname(gid int) string {
	value, _ := gnames.lookup(gname(gid))
	return value
}

// Hostname retrieves and caches host name for ip address.
func Hostname(addr string) string {
	value, err := hnames.lookup(hname(addr))

	if err != nil { // error requests network lookup of hostname
		go func() {
			hnames.Lock()
			defer hnames.Unlock()
			if hs, err := net.LookupAddr(addr); err == nil {
				hnames.values[hname(addr)] = hs[0]
			}
		}()
	}

	return value
}

// Module retrieves and caches go module information.
func Module(dir string) modval {
	value, _ := mnames.lookup(moddir(dir))
	return value
}

// lookup retrieves a user name.
func username(uid uname) (string, error) {
	name := strconv.Itoa(int(uid))
	u, err := user.LookupId(name)
	if err == nil {
		name = u.Name
	}
	return name, nil
}

// groupname retrieves a group name.
func groupname(gid gname) (string, error) {
	name := strconv.Itoa(int(gid))
	g, err := user.LookupGroupId(name)
	if err == nil {
		name = g.Name
	}
	return name, nil
}

// hostname retrieves a host name.
func hostname(addr hname) (string, error) {
	name := string(addr)
	ip, err := netip.ParseAddr(name)
	if err != nil {
		return name, nil
	}
	name = netip.Addr(ip).String() // normalize name
	return name, errors.New("lookup hostname")
}

// modinfo retrieves go module information.
func modinfo(dir moddir) (modval, error) {
	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.NeedName | packages.NeedModule,
			Dir:  string(dir),
		})
	if err != nil {
		return modval{}, nil
	}
	return modval{
		Path: pkgs[0].Module.Path,
		Dir:  pkgs[0].Module.Dir,
		Pkg:  pkgs[0].Name,
	}, nil
}
