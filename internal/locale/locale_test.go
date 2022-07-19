// SPDX-License-Identifier: MIT

package locale

import (
	"encoding/xml"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/internal/serialization"
)

func TestLocale(t *testing.T) {
	a := assert.New(t, false)

	f := serialization.NewFS(5)
	a.NotError(f.Serializer().Add(xml.Marshal, xml.Unmarshal, ".xml"))
	a.NotError(f.Serializer().Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))

	// cmn-hant.xml

	l := New(time.Local, language.MustParse("cmn-hant"))
	a.NotNil(l)
	a.NotError(l.LoadLocaleFiles(os.DirFS("./testdata"), "cmn-hant.xml", f))
	p := l.NewPrinter(language.MustParse("cmn-hant"))

	a.Equal(p.Sprintf("k1"), "msg1")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")

	// cmn-hans.yaml

	l = New(time.Local, language.MustParse("cmn-hans"))
	a.NotNil(l)
	a.NotError(l.LoadLocaleFiles(os.DirFS("./testdata"), "cmn-hans.yaml", f))
	p = l.NewPrinter(language.MustParse("cmn-hans"))

	a.Equal(p.Sprintf("k1"), "msg1")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")
}
