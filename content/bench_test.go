// SPDX-License-Identifier: MIT

package content

import (
	"encoding/xml"
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkBuildContentType(b *testing.B) {
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		a.True(len(buildContentType(DefaultMimetype, DefaultCharset)) > 0)
	}
}

func BenchmarkContent_Marshal(b *testing.B) {
	a := assert.New(b)
	srv := New(DefaultBuilder)
	a.NotNil(srv)

	a.NotError(srv.AddMimetype(xml.Marshal, xml.Unmarshal, "font/wottf"))

	for i := 0; i < b.N; i++ {
		name, marshal, err := srv.marshal("font/wottf;q=0.9")
		a.NotError(err).
			NotEmpty(name).
			NotNil(marshal)
	}
}

func BenchmarkContent_Unmarshal(b *testing.B) {
	a := assert.New(b)
	srv := New(DefaultBuilder)
	a.NotNil(srv)

	a.NotError(srv.AddMimetype(xml.Marshal, xml.Unmarshal, "font/wottf"))

	for i := 0; i < b.N; i++ {
		marshal, err := srv.unmarshal("font/wottf")
		a.NotError(err).
			NotNil(marshal)
	}
}
