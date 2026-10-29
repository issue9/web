// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package locale

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/issue9/assert/v6"
	"github.com/issue9/config"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

func TestLocale_Load(t *testing.T) {
	a := assert.New(t, false)
	s := make(config.Serializer, 2)
	s.Add(xml.Marshal, xml.Unmarshal, ".xml").
		Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")

	b := catalog.NewBuilder()
	a.NotError(Load(s, b, "*.yaml", os.DirFS("./testdata")))

	p1 := message.NewPrinter(language.MustParse("cmn-Hans"), message.Catalog(b))
	a.NotError(b.SetString(language.MustParse("cmn-Hans"), "lang", "hans"))

	a.Equal(p1.Sprintf("k1"), "zh")

	a.Equal(p1.Sprintf("k2", 1), "msg-1")
	a.Equal(p1.Sprintf("k2", 3), "msg-3")
	a.Equal(p1.Sprintf("k2", 5), "msg-other")

	a.Equal(p1.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p1.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p1.Sprintf("k3", 2, 2), "2-二")

	a.NotError(Load(s, b, "*.xml", os.DirFS("./testdata")))
	p2 := message.NewPrinter(language.TraditionalChinese, message.Catalog(b))

	a.Equal(p2.Sprintf("k1"), "zh-hant")

	a.Equal(p2.Sprintf("k2", 1), "msg-1")
	a.Equal(p2.Sprintf("k2", 3), "msg-3")
	a.Equal(p2.Sprintf("k2", 5), "msg-other")

	a.Equal(p2.Sprintf("k3", 1, 1), "1-壹")
	a.Equal(p2.Sprintf("k3", 1, 2), "2-壹")
	a.Equal(p2.Sprintf("k3", 2, 2), "2-贰")
}
