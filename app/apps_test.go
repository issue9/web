// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestNewApps(t *testing.T) {
	a := assert.New(t)

	a.Panic(func() {
		NewApps()
	})

	app1, err := New("./testdata")
	a.NotError(err).NotNil(app1)
	app2, err := New("./testdata")
	a.NotError(err).NotNil(app1)

	// 相同的端口号
	a.Panic(func() {
		NewApps(app1, app2)
	})

	// 改变端口号
	app2.webConfig.Port = 1234
	a.NotPanic(func() {
		NewApps(app1, app2)
	})
}

func TestApps_Serve(t *testing.T) {
	a := assert.New(t)

	app1, err := New("./testdata")
	a.NotError(err).NotNil(app1)
	app2, err := New("./testdata")
	a.NotError(err).NotNil(app1)
	app2.webConfig.Port = 1234

	var apps *Apps
	a.NotPanic(func() {
		apps = NewApps(app1, app2)
	})

	// 启动服务
	go func() {
		a.ErrorType(apps.Serve(), http.ErrServerClosed)
	}()
	time.Sleep(300 * time.Microsecond)

	app1.Close()
	app2.Close()
}
