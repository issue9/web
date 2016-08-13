// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"sync"

	"github.com/issue9/mux"
)

var (
	serveMux = mux.NewServeMux()

	groups   = map[string]*mux.Group{} // 所有模块的列表。
	groupsMu sync.RWMutex
)

// Groups 所有模块列表。
func Groups() map[string]*mux.Group {
	ret := make(map[string]*mux.Group, len(groups))

	groupsMu.RLock()
	for name, group := range groups {
		ret[name] = group
	}
	groupsMu.RUnlock()

	return ret
}

// Group 获取指定名称的 mux.Group，若不存在则返回 nil
func Group(name string) *mux.Group {
	groupsMu.RLock()
	g, _ := groups[name]
	groupsMu.RUnlock()

	return g
}

// MustGroup 获取指定名称的 mux.Group，若不存在则生成新的 mux.Group 实例
func MustGroup(name string) *mux.Group {
	groupsMu.RLock()
	g, found := groups[name]
	groupsMu.RUnlock()

	if !found {
		g = serveMux.Group()
	}

	groupsMu.Lock()
	groups[name] = g
	groupsMu.Unlock()

	return g
}
