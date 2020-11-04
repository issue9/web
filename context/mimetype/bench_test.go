// SPDX-License-Identifier: MIT

package mimetype

import (
	"encoding/xml"
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkServer_Marshal(b *testing.B) {
	a := assert.New(b)
	srv := NewMimetypes()
	a.NotNil(srv)

	a.NotError(srv.AddMarshal("font/wottf", xml.Marshal))

	for i := 0; i < b.N; i++ {
		name, marshal, err := srv.Marshal("font/wottf;q=0.9")
		a.NotError(err).
			NotEmpty(name).
			NotNil(marshal)
	}
}

func BenchmarkServer_Unmarshal(b *testing.B) {
	a := assert.New(b)
	srv := NewMimetypes()
	a.NotNil(srv)

	a.NotError(srv.AddUnmarshal("font/wottf", xml.Unmarshal))

	for i := 0; i < b.N; i++ {
		marshal, err := srv.Unmarshal("font/wottf")
		a.NotError(err).
			NotNil(marshal)
	}
}
