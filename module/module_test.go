// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux"
)

func TestModules_New(t *testing.T) {
	a := assert.New(t)
	ms := NewModules(&mux.Prefix{})
	a.NotNil(ms)

	m, err := ms.New("m1", "m1 desc")
	a.NotError(err).NotNil(m)

	m, err = ms.New("m1", "m1 desc")
	a.ErrorType(err, ErrModuleExists).Nil(m)

	m, err = ms.New("m2", "m1 desc")
	a.NotError(err).NotNil(m)
}

func TestPrefix_Module(t *testing.T) {
	a := assert.New(t)
	ms := NewModules(&mux.Prefix{})
	a.NotNil(ms)

	m, err := ms.New("m1", "m1 desc")
	a.NotError(err).NotNil(m)

	p := m.Prefix("/p")
	a.Equal(p.Module(), m)
}
