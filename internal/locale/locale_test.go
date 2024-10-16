// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package locale

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/config"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"
)

func TestLocale_Printer(t *testing.T) {
	a := assert.New(t, false)

	l := New(language.SimplifiedChinese, nil)
	a.NotError(l.SetString(language.SimplifiedChinese, "lang", "hans")).
		NotNil(l).Equal(l.Sprintf("lang"), "hans").
		NotError(l.SetString(language.SimplifiedChinese, "lang", "hans-2")).
		Equal(l.Sprintf("lang"), "hans-2")

	// ID 不存在于 catalog

	l = New(language.Afrikaans, nil)
	a.NotNil(l).
		NotError(l.SetString(language.SimplifiedChinese, "lang", "hans")).
		Equal(l.Sprintf("lang"), "lang"). // 找不到对应的翻译项，返回原值
		NotError(l.SetString(language.Afrikaans, "lang", "afrik")).
		Equal(l.Sprintf("lang"), "afrik")
}

func TestLocale_NewPrinter(t *testing.T) {
	a := assert.New(t, false)
	l := New(language.SimplifiedChinese, nil)
	a.NotNil(l).Equal(l.ID(), language.SimplifiedChinese)

	// language.SimplifiedChinese 是默认的 ID，初始化 l 时即已存在。

	p1 := l.NewPrinter(language.SimplifiedChinese)
	a.NotError(l.SetString(language.SimplifiedChinese, "lang", "hans"))
	p2 := l.NewPrinter(language.SimplifiedChinese)
	a.Equal(p1.Sprintf("lang"), p2.Sprintf("lang"))

	// language.TraditionalChinese 在调用 SetString 之前不存在，
	// 所以 p1 会匹配成其它相似的值，p2 则会准确匹配到 TraditionalChinese。

	p1 = l.NewPrinter(language.TraditionalChinese)
	a.NotError(l.SetString(language.TraditionalChinese, "lang", "hant"))
	p2 = l.NewPrinter(language.TraditionalChinese)
	a.NotEqual(p1.Sprintf("lang"), p2.Sprintf("lang"))
}

func TestNewPrinter(t *testing.T) {
	a := assert.New(t, false)

	c := catalog.NewBuilder(catalog.Fallback(language.MustParse("zh-TW")))
	a.NotError(c.SetString(language.MustParse("zh-CN"), "k1", "zh-cn")).
		NotError(c.SetString(language.MustParse("zh-TW"), "k1", "zh-tw"))

	p, ok := NewPrinter(language.MustParse("und"), c)
	a.Equal(p.Sprintf("k1"), "zh-tw").False(ok)
}

func Test_Load(t *testing.T) {
	a := assert.New(t, false)

	s := make(config.Serializer, 2)
	s.Add(xml.Marshal, xml.Unmarshal, ".xml")
	s.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
	b := catalog.NewBuilder()
	a.NotError(Load(s, b, "*.*", os.DirFS("./testdata")))

	// zh-hant.xml

	p, exact := NewPrinter(language.MustParse("zh-hant"), b)
	a.True(exact).NotNil(p)

	a.Equal(p.Sprintf("k1"), "zh-hant")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")

	// zh.yaml

	p, exact = NewPrinter(language.MustParse("zh-Hans"), b)
	a.True(exact).NotNil(p)

	a.Equal(p.Sprintf("k1"), "zh")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")

	p, exact = NewPrinter(language.MustParse("cmn-Hans"), b)
	a.True(exact).NotNil(p)
	a.Equal(p.Sprintf("k1"), "zh")

	p, exact = NewPrinter(language.MustParse("zh-cmn-Hans"), b)
	a.True(exact).NotNil(p)
	a.Equal(p.Sprintf("k1"), "zh")

	p, exact = NewPrinter(language.MustParse("zh"), b)
	a.True(exact).NotNil(p)
	a.Equal(p.Sprintf("k1"), "zh")

	p, exact = NewPrinter(language.MustParse("zh-CN"), b)
	a.True(exact).NotNil(p)
	a.Equal(p.Sprintf("k1"), "zh")

	p, exact = NewPrinter(language.MustParse("cmn"), b)
	a.True(exact).NotNil(p)
	a.Equal(p.Sprintf("k1"), "zh")
}
