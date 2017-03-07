// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package request

import (
	"testing"

	"github.com/issue9/assert"
)

func TestParam_Int(t *testing.T) {
	a := assert.New(t)

	data := map[string]string{
		"i":   "-1",
		"str": "str",
	}
	p := newParam(data)
	a.NotNil(p)

	a.Equal(-1, p.Int("i"))
	a.Equal(0, p.Int("str"))
	a.Equal(0, p.Int("not exists"))
	rslt := p.Result(400001)
	a.True(len(rslt.Detail) > 0)

	var i int
	p.IntVar(&i, "i")
	a.Equal(-1, i)

	i = 0
	p.IntVar(&i, "not exists")
	a.Equal(0, i)
}
