// SPDX-License-Identifier: MIT

package files

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"testing"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/internal/errs"
)

type webConfig struct {
	XMLName  struct{} `xml:"web" json:"-"`
	Port     string   `xml:"port,attr" json:"prot"`
	Timezone string   `xml:"timezone" json:"timezone"`
}

func TestFiles_Add_Set_Delete(t *testing.T) {
	a := assert.New(t, false)
	f := New(os.DirFS("./testdata"))

	a.Equal(0, f.Len())
	f.Add(json.Marshal, json.Unmarshal, ".json")
	a.Equal(1, f.Len())

	a.PanicString(func() {
		f.Add(json.Marshal, json.Unmarshal, ".json")
	}, "已经存在同名的扩展名 .json")

	a.PanicString(func() {
		f.Add(json.Marshal, json.Unmarshal)
	}, "参数 ext 不能为空")

	f.Delete(".json")
	a.Equal(0, f.Len())

	f.Add(json.Marshal, json.Unmarshal, ".json", ".js")
	a.Equal(2, f.Len())
}

func TestFiles_Load(t *testing.T) {
	a := assert.New(t, false)
	f := New(os.DirFS("./testdata"))

	web := &webConfig{}
	err := errs.NewLocaleError("not found serialization function for %s", "web.xml")
	a.Equal(f.Load(os.DirFS("./testdata"), "web.xml", web), err)

	f.Set(".xml", xml.Marshal, xml.Unmarshal)
	a.NotError(f.Load(os.DirFS("./testdata"), "web.xml", web))
	a.Equal(web, &webConfig{Port: ":8082", Timezone: "Africa/Addis_Ababa"})
}

func Test_LoadLocales(t *testing.T) {
	a := assert.New(t, false)
	f := New(os.DirFS("./testdata"))

	f.Add(xml.Marshal, xml.Unmarshal, ".xml")
	f.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
	b := catalog.NewBuilder()

	// cmn-hant.xml

	a.NotError(LoadLocales(f, b, os.DirFS("./testdata"), "cmn-*.*"))
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
