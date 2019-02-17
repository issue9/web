// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/issue9/assert"
)

func TestModule_AddInit(t *testing.T) {
	a := assert.New(t)
	m := New(TypeModule, "m1", "m1 desc")
	a.NotNil(m)

	a.Nil(m.Inits)
	m.AddInit(func() error { return nil })
	a.Equal(len(m.Inits), 1).
		NotEmpty(m.Inits[0].Title). // 一个默认的数值。
		NotNil(m.Inits[0].F)

	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.Inits), 2).
		Equal(m.Inits[1].Title, "t1").
		NotNil(m.Inits[1].F)

	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.Inits), 3).
		Equal(m.Inits[2].Title, "t1").
		NotNil(m.Inits[2].F)
}
