// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"bytes"
	"errors"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux"
)

func TestApp_getInit(t *testing.T) {
	a := assert.New(t)
	ms := NewModules(&mux.Prefix{})
	a.NotNil(ms)

	m := ms.New("m1", "m1 desc")
	fn := ms.getInit(m)
	a.NotNil(fn).NotError(fn)

	// 返回错误
	m = ms.New("m2", "m2 desc")
	m.AddInit(func() error {
		return errors.New("error")
	})
	fn = ms.getInit(m)
	a.NotNil(fn).Error(fn())

	w := new(bytes.Buffer)
	m = ms.New("m3", "m3 desc")
	m.AddInit(func() error {
		_, err := w.WriteString("m3")
		return err
	})
	fn = ms.getInit(m)
	a.NotNil(fn).NotError(fn()).Equal(w.String(), "m3")
}

func TestApp_Init(t *testing.T) {
	a := assert.New(t)
	ms := NewModules(&mux.Prefix{})
	a.NotNil(ms)
	w := new(bytes.Buffer)

	m1 := ms.New("m1", "m1", "m2")
	m1.AddInit(func() error {
		_, err := w.WriteString("m1")
		return err
	})
	m2 := ms.New("m2", "m2")
	m2.AddInit(func() error {
		_, err := w.WriteString("m2")
		return err
	})

	a.NotError(ms.Init())
	a.Equal(w.String(), "m2m1")
}
