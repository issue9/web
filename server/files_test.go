// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

type webConfig struct {
	XMLName  struct{} `xml:"web" json:"-"`
	Port     string   `xml:"port,attr" json:"prot"`
	Timezone string   `xml:"timezone" json:"timezone"`
}

func TestFiles_Add_Set_Delete(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	f := s.Files()

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
	s := newServer(a, nil)
	f := s.Files()

	web := &webConfig{}
	err := localeutil.Error("not found serialization function for %s", "web.xml")
	a.Equal(f.Load(os.DirFS("./testdata/files"), "web.xml", web), err)

	f.Set(".xml", xml.Marshal, xml.Unmarshal)
	a.NotError(f.Load(os.DirFS("./testdata/files"), "web.xml", web))
	a.Equal(web, &webConfig{Port: ":8082", Timezone: "Africa/Addis_Ababa"})
}

func TestFiles_LoadLocales(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	f := s.Files()

	f.Add(xml.Marshal, xml.Unmarshal, ".xml")
	f.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")

	// cmn-hant.xml

	a.NotError(f.LoadLocales(os.DirFS("./testdata/files"), "cmn-*.*"))
	p := s.NewPrinter(language.MustParse("cmn-hant"))

	a.Equal(p.Sprintf("k1"), "msg1")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")

	// cmn-hans.yaml

	p = s.NewPrinter(language.MustParse("cmn-hans"))

	a.Equal(p.Sprintf("k1"), "msg1")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")
}
