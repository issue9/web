// SPDX-License-Identifier: MIT

package serialization

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/issue9/assert"
)

type object struct {
	XMLName  struct{} `xml:"web" yaml:"-"`
	Port     string   `xml:"port,attr" yaml:"port"`
	Timezone string   `xml:"timezone" yaml:"timezone"`
}

func TestFiles_Load(t *testing.T) {
	a := assert.New(t)
	f := NewFiles(10)
	a.NotNil(f)
	testdata := os.DirFS("./testdata")

	v := &object{}
	a.ErrorString(f.Load("./testdata/web.xml", v), "未找到适合")

	f.Add(xml.Marshal, xml.Unmarshal, ".xml")
	v = &object{}
	a.NotError(f.LoadFS(testdata, "web.xml", v))
	a.Equal(v.Port, ":8082")

	// 不存在的 yaml
	v = &object{}
	a.Error(f.LoadFS(testdata, "web.yaml", v), "未找到适合")
}

func TestFiles_Save(t *testing.T) {
	a := assert.New(t)
	f := NewFiles(10)
	a.NotNil(f)
	tmp := os.TempDir()

	v := &object{Port: ":333"}
	a.ErrorString(f.Save(tmp+"/web.xml", v), "未找到适合")

	f.Add(xml.Marshal, xml.Unmarshal, ".xml")
	v = &object{Port: ":333"}
	a.NotError(f.Save(tmp+"/web.xml", v))
}
