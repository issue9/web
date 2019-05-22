// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"bytes"
	"testing"

	"github.com/issue9/assert"
)

func TestLogs(t *testing.T) {
	a := assert.New(t)
	initApp(a)

	info := new(bytes.Buffer)
	Info("1,2,3")
	a.Empty(info.String())
	Infof("%d,%d,%d", 1, 2, 3)
	a.Empty(info.String())

	debug := new(bytes.Buffer)
	App().Logs().DEBUG().SetOutput(debug)
	Debug("1,2,3")
	a.Contains(debug.String(), "1,2,3")
	debug.Reset()
	Debugf("%d,%d,%d", 1, 2, 3)
	a.Contains(debug.String(), "1,2,3")

	err := new(bytes.Buffer)
	App().Logs().ERROR().SetOutput(err)
	Error("1,2,3")
	a.Contains(err.String(), "1,2,3")
	err.Reset()
	Errorf("%d,%d,%d", 1, 2, 3)
	a.Contains(err.String(), "1,2,3")

	critical := new(bytes.Buffer)
	App().Logs().CRITICAL().SetOutput(critical)
	Critical("1,2,3")
	a.Contains(critical.String(), "1,2,3")
	critical.Reset()
	Criticalf("%d,%d,%d", 1, 2, 3)
	a.Contains(critical.String(), "1,2,3")

	// Panic
	debug.Reset()
	err.Reset()
	critical.Reset()
	a.Panic(func() {
		Panic("panic!")
	})
	a.Contains(debug.String(), "panic!")
	a.Contains(err.String(), "panic!")
	a.Contains(critical.String(), "panic!")

	// Panicf
	debug.Reset()
	err.Reset()
	critical.Reset()
	a.Panic(func() {
		Panicf("panicf!")
	})
	a.Empty(info.String())
	a.Contains(debug.String(), "panicf!")
	a.Contains(err.String(), "panicf!")
	a.Contains(critical.String(), "panicf!")
}
