// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mimetype

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/mimetype/gob"
)

func BenchmarkMimetypes_Marshal(b *testing.B) {
	a := assert.New(b)
	m := New()
	a.NotNil(m)

	a.NotError(m.AddMarshal(gob.MimeType, gob.Marshal))
	a.NotError(m.AddMarshal("application/json", json.Marshal))
	a.NotError(m.AddMarshal("application/xml", xml.Marshal))
	a.NotError(m.AddMarshal("font/wottf", xml.Marshal))

	for i := 0; i < b.N; i++ {
		name, marshal, err := m.Marshal("font/wottf;q=0.9")
		a.NotError(err).
			NotEmpty(name).
			NotNil(marshal)
	}
}

func BenchmarkMimetypes_Unmarshal(b *testing.B) {
	a := assert.New(b)
	m := New()
	a.NotNil(m)

	a.NotError(m.AddUnmarshal(gob.MimeType, gob.Unmarshal))
	a.NotError(m.AddUnmarshal("application/json", json.Unmarshal))
	a.NotError(m.AddUnmarshal("application/xml", xml.Unmarshal))
	a.NotError(m.AddUnmarshal("font/wottf", xml.Unmarshal))

	for i := 0; i < b.N; i++ {
		marshal, err := m.Unmarshal("font/wottf")
		a.NotError(err).
			NotNil(marshal)
	}
}
