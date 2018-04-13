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

func TestModule_GetInit(t *testing.T) {
	a := assert.New(t)
	router := mux.New(false, false, nil, nil).Prefix("")
	a.NotNil(router)

	m := New("m1", "m1 desc")
	a.NotNil(m)
	fn := m.GetInit(router)
	a.NotNil(fn).NotError(fn())

	// 返回错误
	m = New("m2", "m2 desc")
	a.NotNil(m)
	m.AddInit(func() error {
		return errors.New("error")
	})
	fn = m.GetInit(router)
	a.NotNil(fn).Error(fn())

	w := new(bytes.Buffer)
	m = New("m3", "m3 desc")
	a.NotNil(m)
	m.AddInit(func() error {
		_, err := w.WriteString("m3")
		return err
	})
	m.GetFunc("/get", f1)
	m.Prefix("/p").PostFunc("/post", f1)
	fn = m.GetInit(router)
	a.NotNil(fn).
		NotError(fn()).
		Equal(w.String(), "m3")
}

func TestPrefix_Module(t *testing.T) {
	a := assert.New(t)

	m := New("m1", "m1 desc")
	a.NotNil(m)

	p := m.Prefix("/p")
	a.Equal(p.Module(), m)
}
