// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"encoding/xml"
	"strconv"

	"github.com/issue9/config"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web"
)

var filesFactory = map[string]serializer{}

type MarshalFileFunc = config.MarshalFunc

type UnmarshalFileFunc = config.UnmarshalFunc

type serializer struct {
	Marshal   MarshalFileFunc
	Unmarshal UnmarshalFileFunc
}

func (conf *configOf[T]) sanitizeFileSerializers() *web.FieldError {
	conf.fileSerializers = make(map[string]serializer, len(conf.FileSerializers))
	for i, name := range conf.FileSerializers {
		s, found := filesFactory[name]
		if !found {
			return web.NewFieldError("["+strconv.Itoa(i)+"]", web.NewLocaleError("not found serialization function for %s", name))
		}
		conf.fileSerializers[name] = s // conf.FileSerializers 可以保证 conf.fileSerializers 唯一性
	}
	return nil
}

// RegisterFileSerializer 注册用于文件序列化的方法
//
// ext 为文件的扩展名，如果存在同名，则会覆盖。
func RegisterFileSerializer(m MarshalFileFunc, u UnmarshalFileFunc, ext ...string) {
	for _, e := range ext {
		filesFactory[e] = serializer{Marshal: m, Unmarshal: u}
	}
}

func loadConfigOf[T any](configDir string, name string) (*configOf[T], error) {
	c, err := buildConfig(configDir)
	if err != nil {
		return nil, err
	}

	conf := &configOf[T]{}
	if err := c.Load(name, conf); err != nil {
		return nil, err
	}
	conf.config = c

	return conf, nil
}

func buildConfig(dir string) (*config.Config, error) {
	s := make(config.Serializer, len(filesFactory))
	for ext, item := range filesFactory {
		s.Add(item.Marshal, item.Unmarshal, ext)
	}

	return config.BuildDir(s, dir)
}

func init() {
	RegisterFileSerializer(json.Marshal, json.Unmarshal, ".json")
	RegisterFileSerializer(xml.Marshal, xml.Unmarshal, ".xml")
	RegisterFileSerializer(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
}
