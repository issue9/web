// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestInitSession(t *testing.T) {
	a := assert.New(t)

	cfg.Session = &sessionConfig{
		Type:     "memory",
		IDName:   "gosession",
		Lifetime: 50,
		SaveDir:  "./testdata/",
	}

	sessionMgr = nil
	a.NotPanic(func() { initSession() })
	a.NotNil(sessionMgr)

	sessionMgr = nil
	cfg.Session.Type = "file"
	a.NotPanic(func() { initSession() })
	a.NotNil(sessionMgr)

	sessionMgr = nil
	cfg.Session.Type = "unknown"
	a.Panic(func() { initSession() })
	a.Nil(sessionMgr)
}

func TestSession(t *testing.T) {
	a := assert.New(t)

	cfg.Session = &sessionConfig{
		Type:     "memory",
		IDName:   "gosession",
		Lifetime: 50,
		SaveDir:  "./testdata/",
	}

	sessionMgr = nil
	a.NotPanic(func() { initSession() })
	a.NotNil(sessionMgr)

	req, err := http.NewRequest("GET", "/", nil)
	a.NotError(err).NotNil(req)

	w := httptest.NewRecorder()
	a.NotNil(w)

	sess := Session(w, req)
	a.NotNil(sess)
	a.Equal(len(sessions), 1) // 缓存一个session

	// s应该等同于sess
	s := Session(w, req)
	a.Equal(sess, s)
}
