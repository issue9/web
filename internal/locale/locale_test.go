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

func TestLocale_Printer(t *testing.T) {
	a := assert.New(t, false)

	b := catalog.NewBuilder()
	b.SetString(language.SimplifiedChinese, "lang", "hans")
	l := New(language.SimplifiedChinese, nil, b)
	a.NotNil(l).Equal(l.Sprintf("lang"), "hans")
	l.SetString(language.SimplifiedChinese, "lang", "hans-2")
	a.Equal(l.Sprintf("lang"), "hans-2")

	// ID 不存在于 catalog

	b = catalog.NewBuilder()
	b.SetString(language.SimplifiedChinese, "lang", "hans")
	l = New(language.Afrikaans, nil, b)
	a.NotNil(l).Equal(l.Sprintf("lang"), "lang") // 找不到对应的翻译项，返回原值
	l.SetString(language.Afrikaans, "lang", "afrik")
	a.Equal(l.Sprintf("lang"), "afrik")
}

func TestLocale_NewPrinter(t *testing.T) {
	a := assert.New(t, false)
	l := New(language.SimplifiedChinese, nil, nil)
	a.NotNil(l).Equal(l.ID(), language.SimplifiedChinese)

	// language.SimplifiedChinese 是默认的 ID，初始化 l 时即已存在。

	p1 := l.NewPrinter(language.SimplifiedChinese)
	l.SetString(language.SimplifiedChinese, "lang", "hans")
	p2 := l.NewPrinter(language.SimplifiedChinese)
	a.Equal(p1.Sprintf("lang"), p2.Sprintf("lang"))

	// language.TraditionalChinese 在调用 SetString 之前不存在，
	// 所以 p1 会匹配成其它相似的值，p2 则会准确匹配到 TraditionalChinese。

	p1 = l.NewPrinter(language.TraditionalChinese)
	l.SetString(language.TraditionalChinese, "lang", "hant")
	p2 = l.NewPrinter(language.TraditionalChinese)
	a.NotEqual(p1.Sprintf("lang"), p2.Sprintf("lang"))
}

func TestNewPrinter(t *testing.T) {
	a := assert.New(t, false)

	c := catalog.NewBuilder()
	c.SetString(language.MustParse("zh-CN"), "k1", "zh-cn")
	c.SetString(language.MustParse("zh-TW"), "k1", "zh-tw")

	p := NewPrinter(language.MustParse("cmn-hans"), c)
	a.Equal(p.Sprintf("k1"), "zh-cn")
}

func Test_Load(t *testing.T) {
	a := assert.New(t, false)

	s := make(config.Serializer, 2)
	s.Add(xml.Marshal, xml.Unmarshal, ".xml")
	s.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
	b := catalog.NewBuilder()
	a.NotError(Load(s, b, "cmn-*.*", os.DirFS("./testdata")))

	// cmn-hant.xml

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
