// SPDX-License-Identifier: MIT

package filesystem

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/localeutil"

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
