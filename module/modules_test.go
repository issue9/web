// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux"
)

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNewModules(t *testing.T) {
	a := assert.New(t)

	a.Panic(func() {
		NewModules(nil) // 空参数 router
	})

	ms := NewModules(&mux.Prefix{})
	a.NotNil(ms)
}

func TestModules_Modules(t *testing.T) {
	a := assert.New(t)
	ms := NewModules(&mux.Prefix{})
	a.NotNil(ms)

	_, err := ms.New("m1", "m1 desc")
	a.NotError(err)
	list := ms.Modules()
	a.Equal(len(list), 1)

	_, err = ms.New("m2", "m1 desc")
	a.NotError(err)
	list = ms.Modules()
	a.Equal(len(list), 2)
}

func TestModules_getInit(t *testing.T) {
	a := assert.New(t)
	ms := NewModules(mux.New(false, false, nil, nil).Prefix(""))
	a.NotNil(ms)

	m, err := ms.New("m1", "m1 desc")
	a.NotError(err).NotNil(m)
	fn := ms.getInit(m)
	a.NotNil(fn).NotError(fn)

	// 返回错误
	m, err = ms.New("m2", "m2 desc")
	a.NotError(err).NotNil(m)
	m.AddInit(func() error {
		return errors.New("error")
	})
	fn = ms.getInit(m)
	a.NotNil(fn).Error(fn())

	w := new(bytes.Buffer)
	m, err = ms.New("m3", "m3 desc")
	a.NotError(err).NotNil(m)
	m.AddInit(func() error {
		_, err := w.WriteString("m3")
		return err
	})
	m.GetFunc("/get", f1)
	m.Prefix("/p").PostFunc("/post", f1)
	fn = ms.getInit(m)
	a.NotNil(fn).
		NotError(fn()).
		Equal(w.String(), "m3")
}

func TestModules_Init(t *testing.T) {
	a := assert.New(t)
	ms := NewModules(&mux.Prefix{})
	a.NotNil(ms)
	w := new(bytes.Buffer)

	m1, err := ms.New("m1", "m1", "m2")
	a.NotError(err).NotNil(m1)
	m1.AddInit(func() error {
		_, err := w.WriteString("m1")
		return err
	})
	m2, err := ms.New("m2", "m2")
	a.NotError(err).NotNil(m2)
	m2.AddInit(func() error {
		_, err := w.WriteString("m2")
		return err
	})

	a.NotError(ms.Init())
	a.Equal(w.String(), "m2m1")

	// 多次初始化
	a.ErrorType(ms.Init(), ErrModulesIsInited)
}
