// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

type empty struct{}

func TestApp_init(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		app := &App{}
		app.init()
	}, "app.NewServer 不能为空")

	app := &App{NewServer: func() (web.Server, error) {
		return server.New("test", "1.0.0", nil)
	}}
	a.NotError(app.init())
}
