// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accept

import (
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkParseAccept(b *testing.B) {
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		_, _, err := parseAccept("application/xml;q=0.9")
		a.NotError(err)
	}
}

func BenchmarkParse_mult(b *testing.B) {
	a := assert.New(b)

	str := "application/json;q=0.9,text/plain;q=0.8,text/html,text/xml,*;q=0.1"
	for i := 0; i < b.N; i++ {
		as, err := Parse(str)
		a.NotError(err).True(len(as) > 0)
	}
}

func BenchmarkParse_one(b *testing.B) {
	a := assert.New(b)

	str := "application/json;q=0.9"
	for i := 0; i < b.N; i++ {
		as, err := Parse(str)
		a.NotError(err).True(len(as) > 0)
	}
}

func TestParseAccept(t *testing.T) {
	a := assert.New(t)

	v, q, err := parseAccept("application/xml")
	a.NotError(err).
		Equal(v, "application/xml").
		Equal(q, 1.0)

	v, q, err = parseAccept("application/xml;")
	a.NotError(err).
		Equal(v, "application/xml").
		Equal(q, 1.0)

	v, q, err = parseAccept("application/xml;q=0.9")
	a.NotError(err).
		Equal(v, "application/xml").
		Equal(q, float32(0.9))

	v, q, err = parseAccept("text/html;format=xx;q=0.9")
	a.NotError(err).
		Equal(v, "text/html").
		Equal(q, float32(0.9))

	v, q, err = parseAccept("text/html;format=xx;q=x.9")
	a.Error(err).Empty(v).Empty(q)

	v, q, err = parseAccept("text/html;format=xx;q=0.9x")
	a.Error(err).Empty(v).Empty(q)
}

func TestParse(t *testing.T) {
	a := assert.New(t)

	as, err := Parse(",a1,a2,a3;q=0.5,a4,a5;q=0.9,a6;a61;q=0.8")
	a.NotError(err).NotEmpty(as)
	a.Equal(len(as), 6)
	// 确定排序是否正常
	a.Equal(as[0].Q, float32(1.0))
	a.Equal(as[5].Q, float32(.5))

	as, err = Parse(",a1,a2,a3;q=0.5,a4,a5;q=0.9,a6;a61;q=0.0")
	a.NotError(err).NotEmpty(as)
	a.Equal(len(as), 5)
	a.Equal(as[0].Q, float32(1.0))

	as, err = Parse(",a1,a2,a3;q=5,a4,a5;q=0.9,a6;a61;q=0.x8")
	a.Error(err).Empty(as)

	as, err = Parse("utf-8;q=x.9,gbk;q=0.8")
	a.Error(err).Empty(as)
}
