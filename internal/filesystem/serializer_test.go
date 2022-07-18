// SPDX-License-Identifier: MIT

package filesystem

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/serializer"
)

var _ serializer.FS = &Serializer{}

type object struct {
	XMLName  struct{} `xml:"web" yaml:"-"`
	Port     string   `xml:"port,attr" yaml:"port"`
	Timezone string   `xml:"timezone" yaml:"timezone"`
}

func TestSerializer_Load(t *testing.T) {
	a := assert.New(t, false)
	f := NewSerializer(serializer.New(5))
	a.NotNil(f)
	testdata := os.DirFS("./testdata")

	v := &object{}
	a.Equal(f.Load(os.DirFS("./testdata"), "web.xml", v), localeutil.Error("not found serialization function for %s", "web.xml"))

	a.NotError(f.Serializer().Add(xml.Marshal, xml.Unmarshal, ".xml"))
	v = &object{}
	a.NotError(f.Load(testdata, "web.xml", v))
	a.Equal(v.Port, ":8082")

	// 不存在的 yaml
	v = &object{}
	a.Error(f.Load(testdata, "web.yaml", v))
}

func TestSerializer_Save(t *testing.T) {
	a := assert.New(t, false)
	f := NewSerializer(serializer.New(10))
	a.NotNil(f)
	tmp := os.TempDir()

	v := &object{Port: ":333"}
	a.Equal(f.Save(tmp+"/web.xml", v), localeutil.Error("not found serialization function for %s", tmp+"/web.xml"))

	a.NotError(f.Serializer().Add(xml.Marshal, xml.Unmarshal, ".xml"))
	v = &object{Port: ":333"}
	a.NotError(f.Save(tmp+"/web.xml", v))
}

func TestLoadLocaleFiles(t *testing.T) {
	a := assert.New(t, false)
	f := NewSerializer(serializer.New(5))
	b := catalog.NewBuilder()
	a.NotNil(b)

	a.NotError(f.Serializer().Add(xml.Marshal, xml.Unmarshal, ".xml"))
	a.NotError(f.Serializer().Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))

	a.NotError(LoadLocaleFiles(f, b, os.DirFS("./testdata"), "cmn-hant.xml"))
	p := message.NewPrinter(language.MustParse("cmn-hant"), message.Catalog(b))

	a.Equal(p.Sprintf("k1"), "msg1")

	a.Equal(p.Sprintf("k2", 1), "msg-1")
	a.Equal(p.Sprintf("k2", 3), "msg-3")
	a.Equal(p.Sprintf("k2", 5), "msg-other")

	a.Equal(p.Sprintf("k3", 1, 1), "1-一")
	a.Equal(p.Sprintf("k3", 1, 2), "2-一")
	a.Equal(p.Sprintf("k3", 2, 2), "2-二")
}
