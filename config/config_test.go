// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"testing"

	"github.com/issue9/assert"
)

type web struct {
	XMLName  struct{} `yaml:"-" json:"-" xml:"webconfig"`
	Debug    bool     `yaml:"debug,omitempty" json:"debug,omitempty" xml:"debug,attr,omitempty"`
	Root     string   `yaml:"root,omitempty" json:"root,omitempty" xml:"root,omitempty"`
	Timezone string   `yaml:"timezone,omitempty" json:"timezone,omitempty" xml:"timezone,omitempty"`
}

func testUnmarshal(config string, v interface{}) error {
	v.(*web).Debug = !v.(*web).Debug
	return nil
}

func TestConfig_Register(t *testing.T) {
	a := assert.New(t)

	cfg := &Config{}
	w := &web{}

	// 可以使用多个空的 id 参数
	a.NotError(cfg.Register("", "./testdata/web.yaml", w, LoadYAML, nil))
	a.NotError(cfg.Register("", "./testdata/web.yaml", w, LoadYAML, nil))

	// 不可以使用多个非空的 id 参数
	a.NotError(cfg.Register("id1", "./testdata/web.yaml", w, LoadYAML, nil))
	a.Error(cfg.Register("id1", "./testdata/web.yaml", w, LoadYAML, nil))

	a.NotError(cfg.Register("id2", "./testdata/web.yaml", w, LoadYAML, nil))
	a.NotError(cfg.Register("id3", "./testdata/web.yaml", w, LoadYAML, nil))
}

func TestConfig_Refresh(t *testing.T) {
	a := assert.New(t)

	cfg := &Config{}
	w := &web{}
	var cnt int

	a.NotError(cfg.Register("", "./testdata/web.yaml", w, LoadYAML, nil))
	a.NotError(cfg.Register("test", "test", w, testUnmarshal, func() { cnt++ }))

	// 参数 ID 不能为空
	a.Panic(func() {
		cfg.Refresh("")
	})

	// 不存在的 ID
	a.Equal(cfg.Refresh("not-exists"), ErrNotFound)

	a.NotError(cfg.Refresh("test"))
	a.True(w.Debug).Equal(1, cnt)

	a.NotError(cfg.Refresh("test"))
	a.False(w.Debug).Equal(2, cnt)
}

func TestLoadYAML(t *testing.T) {
	a := assert.New(t)

	conf := &web{}
	a.NotError(LoadYAML("./testdata/web.yaml", conf))
	a.True(conf.Debug).
		Equal(conf.Root, "http://localhost:8082")

	a.Error(LoadYAML("./testdata/web.xml", conf))
	a.ErrorIs(LoadYAML("./testdata/not-exists", conf), os.ErrNotExist)
}

func TestLoadJSON(t *testing.T) {
	a := assert.New(t)

	conf := &web{}
	a.NotError(LoadJSON("./testdata/web.json", conf))
	a.True(conf.Debug).
		Equal(conf.Root, "http://localhost:8082")

	a.Error(LoadYAML("./testdata/web.xml", conf))
	a.ErrorIs(LoadJSON("./testdata/not-exists", conf), os.ErrNotExist)
}

func TestLoadXML(t *testing.T) {
	a := assert.New(t)

	conf := &web{}
	a.NotError(LoadXML("./testdata/web.xml", conf))
	a.True(conf.Debug).
		Equal(conf.Root, "http://localhost:8082")

	a.Error(LoadYAML("./testdata/web.xml", conf))
	a.ErrorIs(LoadXML("./testdata/not-exists", conf), os.ErrNotExist)
}
