// SPDX-License-Identifier: MIT

package html

import (
	"html/template"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/serializer"
)

var (
	_ serializer.MarshalFunc   = Marshal
	_ serializer.UnmarshalFunc = Unmarshal
)

func TestNewTpl(t *testing.T) {
	a := assert.New(t, false)
	tt, err := template.ParseGlob("./testdata/*.tpl")
	a.NotError(err).NotNil(tt)

	obj := &struct{}{}
	tpl := NewTpl(tt, "n1", obj)
	a.NotNil(tpl).
		Equal(tpl.name, "n1").
		Equal(tpl.tpl, tt).
		Equal(tpl.data, obj)

	tpl = NewTpl(tt, "n2", tpl)
	a.NotNil(tpl).
		Equal(tpl.name, "n2").
		Equal(tpl.tpl, tt).
		Equal(tpl.data, obj)
}

func TestMarshal(t *testing.T) {
	a := assert.New(t, false)

	tpl, err := template.ParseGlob("./testdata/*.tpl")
	a.NotError(err).NotNil(tpl)

	bs, err := Marshal(NewTpl(tpl, "footer", map[string]any{
		"Footer": "footer",
	}))
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), "<div>footer</div>")

	bs, err = Marshal(NewTpl(tpl, "header", &struct{ Header string }{
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
