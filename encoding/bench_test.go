// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkBuildContentType(b *testing.B) {
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		a.True(len(BuildContentType(DefaultMimeType, utf8Name)) > 0)
	}
}

func BenchmarkParseContentType(b *testing.B) {
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		_, _, err := ParseContentType("application/json;charset=utf-8")
		a.NotError(err)
	}
}
