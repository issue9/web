// SPDX-License-Identifier: MIT

package locale

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/config"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"
)

func Test_Load(t *testing.T) {
	a := assert.New(t, false)

	s := make(config.Serializer, 2)
	s.Add(xml.Marshal, xml.Unmarshal, ".xml")
	s.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
	b := catalog.NewBuilder()

	// cmn-hant.xml

	a.NotError(Load(s, b, os.DirFS("./testdata"), "cmn-*.*"))
	p := message.NewPrinter(language.MustParse("cmn-hant"), message.Catalog(b))

	a.Equal(p.Sprintf("k1"), "msg1")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")

	// cmn-hans.yaml

	p = message.NewPrinter(language.MustParse("cmn-hans"), message.Catalog(b))

	a.Equal(p.Sprintf("k1"), "msg1")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")
}

func Test_LoadGlob(t *testing.T) {
	a := assert.New(t, false)

	s := make(config.Serializer, 2)
	s.Add(xml.Marshal, xml.Unmarshal, ".xml")
	s.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
	b := catalog.NewBuilder()

	// cmn-hant.xml

	a.NotError(LoadGlob(s, b, "./testdata/cmn-*.*"))
	p := message.NewPrinter(language.MustParse("cmn-hant"), message.Catalog(b))

	a.Equal(p.Sprintf("k1"), "msg1")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")

	// cmn-hans.yaml

	p = message.NewPrinter(language.MustParse("cmn-hans"), message.Catalog(b))

	a.Equal(p.Sprintf("k1"), "msg1")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")
}
