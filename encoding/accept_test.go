// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"encoding/json"
	"testing"

	"github.com/issue9/assert"
)

func TestAccept_parse(t *testing.T) {
	a := assert.New(t)

	accept := &Accept{Content: "application/xml"}
	a.NotError(accept.parse())
	a.Equal(accept.Value, accept.Content).
		Equal(accept.Q, 1.0)

	accept = &Accept{Content: "application/xml;"}
	a.NotError(accept.parse())
	a.Equal(accept.Value, "application/xml").
		Equal(accept.Q, 1.0)

	accept = &Accept{Content: "application/xml;q=0.9"}
	a.NotError(accept.parse())
	a.Equal(accept.Value, "application/xml").
		Equal(accept.Q, float32(0.9))

	accept = &Accept{Content: "text/html;format=xx;q=0.9"}
	a.NotError(accept.parse())
	a.Equal(accept.Value, "text/html").
		Equal(accept.Q, float32(0.9))

	accept = &Accept{Content: "text/html;format=xx;q=0.9x"}
	a.Error(accept.parse())
}

func TestAcceptMimeType(t *testing.T) {
	a := assert.New(t)

	name, marshal, err := AcceptMimeType(DefaultMimeType, "", nil)
	a.NotError(err).NotNil(marshal).NotEmpty(name)

	name, marshal, err = AcceptMimeType("font/wotff", "wtoff", json.Marshal)
	a.NotError(err).
		Equal(name, "wtoff").
		Equal(marshal, MarshalFunc(json.Marshal))
}

func TestParseAccept(t *testing.T) {
	a := assert.New(t)

	as, err := ParseAccept(",a1,a2,a3;q=0.5,a4,a5=0.9,a6;a61;q=0.8")
	a.NotError(err).NotEmpty(as)
	a.Equal(len(as), 6)
	// 确定排序是否正常
	a.Equal(as[0].Q, float32(1.0))
	a.Equal(as[5].Q, float32(.5))
}
