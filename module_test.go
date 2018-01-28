// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"bytes"
	"errors"
	"testing"

	"github.com/issue9/assert"
)

func TestApp_getInit(t *testing.T) {
	a := assert.New(t)
	app, err := newApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	m := NewModule("m1", "m1 desc")
	app.AddModule(m)
	fn := app.getInit(m)
	a.NotNil(fn).NotError(fn)

	// 返回错误
	m = NewModule("m2", "m2 desc")
	m.AddInit(func() error {
		return errors.New("error")
	})
	fn = app.getInit(m)
	a.NotNil(fn).Error(fn())

	w := new(bytes.Buffer)
	m = NewModule("m3", "m3 desc")
	m.AddInit(func() error {
		_, err := w.WriteString("m3")
		return err
	})
	fn = app.getInit(m)
	a.NotNil(fn).NotError(fn()).Equal(w.String(), "m3")
}

func TestApp_Modules(t *testing.T) {
	a := assert.New(t)
	app, err := newApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	app.AddModule(NewModule("m1", "m1", "m2"))
	app.AddModule(NewModule("m2", "m2"))

	a.Equal(2, len(app.Modules()))
}

func TestApp_initDependency(t *testing.T) {
	a := assert.New(t)
	app, err := newApp("./testdata", nil)
	a.NotError(err).NotNil(app)
	w := new(bytes.Buffer)

	m1 := NewModule("m1", "m1", "m2")
	m1.AddInit(func() error {
		_, err := w.WriteString("m1")
		return err
	})
	m2 := NewModule("m2", "m2")
	m2.AddInit(func() error {
		_, err := w.WriteString("m2")
		return err
	})

	app.AddModule(m1)
	app.AddModule(m2)
	a.NotError(app.initDependency())
	a.Equal(w.String(), "m2m1")
}
