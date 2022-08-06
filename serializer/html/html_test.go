// SPDX-License-Identifier: MIT

package html

import (
	"html/template"
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/web/serializer"
)

var (
	_ serializer.MarshalFunc   = Marshal
	_ serializer.UnmarshalFunc = Unmarshal
)

func TestMarshal(t *testing.T) {
	a := assert.New(t, false)

	tpl, err := template.ParseGlob("./testdata/*.tpl")
	a.NotError(err).NotNil(tpl)

	bs, err := Marshal(Tpl(tpl, "footer", map[string]any{
		"Footer": "footer",
	}))
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), "<div>footer</div>")

	bs, err = Marshal(Tpl(tpl, "header", &struct{ Header string }{
		Header: "header",
	}))
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), "<div>header</div>")

	bs, err = Marshal(5)
	a.Error(err).Nil(bs)

	bs, err = Marshal("<div>abc</div>")
	a.NotError(err).Equal(string(bs), "<div>abc</div>")

	bs, err = Marshal([]byte("<div>abc</div>"))
	a.NotError(err).Equal(string(bs), "<div>abc</div>")
}
