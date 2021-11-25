// SPDX-License-Identifier: MIT

package serialization

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"
)

func TestLocale_LoadFile(t *testing.T) {
	a := assert.New(t, false)
	f := NewFiles(10)
	l := NewLocale(catalog.NewBuilder(), f)
	a.NotNil(l)

	a.Error(l.LoadFile("./testdata/*.yaml"))

	a.NotError(f.Add(xml.Marshal, xml.Unmarshal, ".xml"))
	a.NotError(f.Add(yaml.Marshal, yaml.Unmarshal, ".yaml"))

	a.NotError(l.LoadFile("./testdata/*.yaml"))
	p := l.Printer(language.MustParse("cmn-hans"))

	a.Equal(p.Sprintf("k1"), "msg1")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")
}

func TestLocale_LoadFileFS(t *testing.T) {
	a := assert.New(t, false)
	f := NewFiles(10)
	l := NewLocale(catalog.NewBuilder(), f)
	a.NotNil(l)

	a.NotError(f.Add(xml.Marshal, xml.Unmarshal, ".xml"))
	a.NotError(f.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))

	a.NotError(l.LoadFileFS(os.DirFS("./testdata"), "cmn-hant.xml"))
	p := l.Printer(language.MustParse("cmn-hant"))

	a.Equal(p.Sprintf("k1"), "msg1")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")
}
