// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestQuery_Int(t *testing.T) {
	a := assert.New(t)

	form := map[string][]string{
		"i":   []string{"-1"},
		"str": []string{"str"},
	}

	r := &http.Request{Form: form}
	ctx := NewContext(nil, r)
	q := ctx.Queries()
	a.Equal(-1, q.Int("i", 2))
	a.Equal(2, q.Int("str", 2))
	a.Equal(2, q.Int("not exists", 2))
	rslt := q.Result(400001)
	a.True(len(rslt.Detail) > 0)

	var i int
	q.IntVar(&i, "i", 2)
	a.Equal(-1, i)
	q.IntVar(&i, "not exists", 2)
	a.Equal(2, i)
}
