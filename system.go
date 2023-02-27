// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"net"
	"net/netip"
	"os/user"
	"runtime"
	"strconv"
	"sync"
	"time"

	"golang.org/x/tools/go/packages"
)

type (
	// name is generic type for cached value.
	name[V any] interface {
		comparable
		lookup() V
	}

	// cache defines a type for mapping ids to names.
	cache[K name[V], V any] struct {
		sync.Mutex
		names map[K]V
	}

	// uname is key to cached user name.
	uname int

	// gname is key to cached group name.
	gname int

	// hname is key to cached host name.
	hname netip.Addr

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
	unames = cache[uname, string]{names: map[uname]string{}}

	// gnames is the cache of group names.
	gnames = cache[gname, string]{names: map[gname]string{}}

	// hnames is the cache of host names.
	hnames = cache[hname, string]{names: map[hname]string{}}

	// mnames is the cache of go module information.
	mnames = cache[moddir, modval]{names: map[moddir]modval{}}
)

// Username retrieves and caches user name for uid.
func Username(uid int) string {
	return lookup(uname(uid), &unames)
}

// Groupname retrieves and caches group name for gid.
func Groupname(gid int) string {
	return lookup(gname(gid), &gnames)
}

// Hostname retrieves and caches host name for ip address.
func Hostname(addr string) string {
	ip, err := netip.ParseAddr(addr)
	if err != nil {
		return addr
	}
	return lookup(hname(ip), &hnames)
}

// Module retrieves and caches go module information.
func Module(dir string) modval {
	return lookup(moddir(dir), &mnames)
}

// lookup looks up a cached value by key.
func lookup[K name[V], V any](key K, names *cache[K, V]) V {
	names.Lock()
	defer names.Unlock()

	if val, ok := names.names[key]; ok {
		return val
	}

	val := key.lookup()
	names.names[key] = val
	return val
}

// lookup retrieves a user name.
func (uid uname) lookup() string {
	name := strconv.Itoa(int(uid))
	if u, err := user.LookupId(name); err == nil {
		name = u.Name
	}
	return name
}

// lookup retrieves a group name.
func (gid gname) lookup() string {
	name := strconv.Itoa(int(gid))
	if g, err := user.LookupGroupId(name); err == nil {
		name = g.Name
	}
	return name
}

// lookup retrieves a host name.
func (addr hname) lookup() string {
	ip := netip.Addr(addr).String()
	go func() { // initiate host name lookup
		if hs, err := net.LookupAddr(ip); err == nil {
			hnames.Lock()
			hnames.names[addr] = hs[0]
			hnames.Unlock()
		}
	}()
	return ip
}

// lookup retrieves go module information.
func (dir moddir) lookup() modval {
	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.NeedName | packages.NeedModule,
			Dir:  string(dir),
		})
	if err != nil {
		return modval{}
	}
	return modval{
		Path: pkgs[0].Module.Path,
		Dir:  pkgs[0].Module.Dir,
		Pkg:  pkgs[0].Name,
	}
}

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
