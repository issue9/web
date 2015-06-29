// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"sync"

	"github.com/issue9/session"
	"github.com/issue9/session/providers"
	"github.com/issue9/session/stores"
)

var (
	sessionMgr *session.Manager
	sessions   = map[*http.Request]*session.Session{}
	sessionsMu sync.Mutex
)

// 初始化session
// vars接受以下四个参数：idname,lifetime,type,saveDir
func initSession() {
	c := cfg.Session
	prv := providers.NewCookie(c.Lifetime, c.IDName, "/", "", true)

	switch c.Type {
	case "", "memory":
		sessionMgr = session.New(stores.NewMemory(c.Lifetime), prv)
	case "file":
		f, err := stores.NewFile(c.SaveDir, c.Lifetime, ERROR())
		if err != nil {
			panic(err)
		}
		sessionMgr = session.New(f, prv)
	default:
		panic("initSession:无效的session.type值")
	}
}

// 获取Session实例。
func Session(w http.ResponseWriter, req *http.Request) *session.Session {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if sess, found := sessions[req]; found {
		return sess
	}

	sess, err := sessionMgr.Start(w, req)
	if err != nil {
		Error(err)
		return nil
	}

	sessions[req] = sess
	return sess
}
