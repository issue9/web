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
)

func TestLocale_Sprintf(t *testing.T) {
	a := assert.New(t, false)

	l := New(language.SimplifiedChinese, nil)
	a.NotError(l.Catalog().SetString(language.SimplifiedChinese, "lang", "hans")).
		NotNil(l).Equal(l.Sprintf("lang"), "hans").
		NotError(l.Catalog().SetString(language.SimplifiedChinese, "lang", "hans-2")).
		Equal(l.Sprintf("lang"), "hans-2")

	// ID 不存在于 catalog

	l = New(language.Afrikaans, nil)
	a.NotNil(l).
		NotError(l.Catalog().SetString(language.SimplifiedChinese, "lang", "hans")).
		Equal(l.Sprintf("lang"), "lang"). // 找不到对应的翻译项，返回原值
		NotError(l.Catalog().SetString(language.Afrikaans, "lang", "afrik")).
		Equal(l.Sprintf("lang"), "afrik")
}

func TestLocale_AcceptLanguage(t *testing.T) {
	a := assert.New(t, false)
	l := New(language.SimplifiedChinese, nil)

	b := l.Catalog()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))
	a.NotError(b.SetString(language.AmericanEnglish, "lang", "en_US"))

	tag := l.AcceptLanguage("")
	a.Equal(tag, l.ID(), "v1:%s, v2:%s", tag.String(), language.Und.String())

	tag = l.AcceptLanguage("zh") // 匹配 zh-hans
	a.Equal(tag, language.SimplifiedChinese, "v1:%s, v2:%s", tag.String(), language.SimplifiedChinese.String())

	tag = l.AcceptLanguage("zh-Hant")
	a.Equal(tag, language.TraditionalChinese, "v1:%s, v2:%s", tag.String(), language.TraditionalChinese.String())

	tag = l.AcceptLanguage("zh-Hans")
	a.Equal(tag, language.SimplifiedChinese, "v1:%s, v2:%s", tag.String(), language.SimplifiedChinese.String())

	tag = l.AcceptLanguage("english") // english 非正确的 tag，但是常用。
	a.Equal(tag, language.AmericanEnglish, "v1:%s, v2:%s", tag.String(), language.AmericanEnglish.String())

	tag = l.AcceptLanguage("zh-Hans;q=0.1,zh-Hant;q=0.3,en")
	a.Equal(tag, language.AmericanEnglish, "v1:%s, v2:%s", tag.String(), language.AmericanEnglish.String())
}

func TestLocale_AcceptLanguageHeader(t *testing.T) {
	a := assert.New(t, false)
	s := make(config.Serializer, 2)
	s.Add(xml.Marshal, xml.Unmarshal, ".xml")
	s.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")

	l := New(language.SimplifiedChinese, config.Dir(s, "./"))
	a.StrictEqual(l.AcceptLanguageHeader(), "zh-Hans")

	a.NotError(l.LoadMessages("*.yaml", os.DirFS("./testdata"))).
		StrictEqual(l.AcceptLanguageHeader(), "zh-Hans, cmn-Hans")

	a.NotError(l.LoadMessages("*.xml", os.DirFS("./testdata"))).
		StrictEqual(l.AcceptLanguageHeader(), "zh-Hans, cmn-Hans, zh-Hant")
}

func TestLocale_NewPrinter(t *testing.T) {
	a := assert.New(t, false)
	s := make(config.Serializer, 2)
	s.Add(xml.Marshal, xml.Unmarshal, ".xml").
		Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")

	l := New(language.MustParse("cmn-Hans"), config.Dir(s, "./"))
	a.NotNil(l).Equal(l.ID(), language.MustParse("cmn-Hans"))

	// language.MustParse("cmn-Hans") 是默认的 ID，初始化 l 时即已存在。

	p1 := l.NewPrinter(language.MustParse("cmn-Hans"))
	a.NotError(l.Catalog().SetString(language.MustParse("cmn-Hans"), "lang", "hans"))

	a.NotError(l.LoadMessages("*.yaml", os.DirFS("./testdata")))
	p2 := l.NewPrinter(language.MustParse("cmn-Hans"))
	a.Equal(p1.Sprintf("lang"), p2.Sprintf("lang"))
	a.Equal(p2.Sprintf("k1"), "zh")

	a.Equal(p2.Sprintf("k2", 1), "msg-1")
	a.Equal(p2.Sprintf("k2", 3), "msg-3")
	a.Equal(p2.Sprintf("k2", 5), "msg-other")

	a.Equal(p2.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p2.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p2.Sprintf("k3", 2, 2), "2-二")

	// language.TraditionalChinese 在调用 SetString 之前不存在，
	// 所以 p1 会匹配成其它相似的值，p2 则会准确匹配到 TraditionalChinese。

	p1 = l.NewPrinter(language.TraditionalChinese)
	a.NotError(l.Catalog().SetString(language.TraditionalChinese, "lang", "hant"))

	a.NotError(l.LoadMessages("*.xml", os.DirFS("./testdata")))
	p2 = l.NewPrinter(language.TraditionalChinese)
	a.NotEqual(p1.Sprintf("lang"), p2.Sprintf("lang"))

	a.Equal(p2.Sprintf("k1"), "zh-hant")

	a.Equal(p2.Sprintf("k2", 1), "msg-1")
	a.Equal(p2.Sprintf("k2", 3), "msg-3")
	a.Equal(p2.Sprintf("k2", 5), "msg-other")

	a.Equal(p2.Sprintf("k3", 1, 1), "1-壹")
	a.Equal(p2.Sprintf("k3", 1, 2), "2-壹")
	a.Equal(p2.Sprintf("k3", 2, 2), "2-贰")
}
