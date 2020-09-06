// SPDX-License-Identifier: MIT

package html

import (
	"html/template"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/context/mimetype"
)

var _ mimetype.MarshalFunc = (&HTML{}).Marshal

func TestHTML(t *testing.T) {
	a := assert.New(t)

	tpl, err := template.ParseGlob("./testdata/*.tpl")
	a.NotError(err).NotNil(tpl)

	mgr := New(tpl)
	bs, err := mgr.Marshal(Tpl("footer", map[string]interface{}{
		"Footer": "footer",
	}))
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), "<div>footer</div>")

	bs, err = mgr.Marshal(Tpl("header", &struct{ Header string }{
		Header: "header",
	}))
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), "<div>header</div>")

	bs, err = mgr.Marshal(5)
	a.Error(err).Nil(bs)
}
