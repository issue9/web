// SPDX-License-Identifier: MIT

package content

import (
	"encoding/xml"
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkParseContentType(b *testing.B) {
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		_, _, err := ParseContentType("appliCation/json;Charset=utf-8")
		a.NotError(err)
	}
}

func BenchmarkBuildContentType(b *testing.B) {
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		a.True(len(BuildContentType(DefaultMimetype, DefaultCharset)) > 0)
	}
}

func BenchmarkMimetypes_Marshal(b *testing.B) {
	a := assert.New(b)
	srv := NewMimetypes()
	a.NotNil(srv)

	a.NotError(srv.Add("font/wottf", xml.Marshal, xml.Unmarshal))

	for i := 0; i < b.N; i++ {
		name, marshal, err := srv.Marshal("font/wottf;q=0.9")
		a.NotError(err).
			NotEmpty(name).
			NotNil(marshal)
	}
}

func BenchmarkMimetypes_Unmarshal(b *testing.B) {
	a := assert.New(b)
	srv := NewMimetypes()
	a.NotNil(srv)

	a.NotError(srv.Add("font/wottf", xml.Marshal, xml.Unmarshal))

	for i := 0; i < b.N; i++ {
		marshal, err := srv.Unmarshal("font/wottf")
		a.NotError(err).
			NotNil(marshal)
	}
}
