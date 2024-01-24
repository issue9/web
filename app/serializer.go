// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"slices"
	"strconv"

	"github.com/issue9/config"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

var filesFactory = newRegister[*server.FileSerializer]()

func (conf *configOf[T]) sanitizeFileSerializers() *web.FieldError {
	for i, name := range conf.FileSerializers {
		s, found := filesFactory.get(name)
		if !found {
			return web.NewFieldError("["+strconv.Itoa(i)+"]", web.NewLocaleError("not found serialization function for %s", name))
		}
		conf.config.Serializers = append(conf.config.Serializers, s)
	}
	return nil
}

// RegisterFileSerializer 注册用于文件序列化的方法
//
// ext 为文件的扩展名；
// name 为当前数据的名称，如果存在同名，则会覆盖；
func RegisterFileSerializer(name string, m config.MarshalFunc, u config.UnmarshalFunc, ext ...string) {
	for _, e := range ext {
		for k, s := range filesFactory.items {
			if slices.Index(s.Exts, e) >= 0 {
				panic(fmt.Sprintf("扩展名 %s 已经注册到 %s", e, k))
			}
		}
	}
	filesFactory.register(&server.FileSerializer{Marshal: m, Unmarshal: u, Exts: ext}, name)
}

func loadConfigOf[T any](configDir, name string) (*configOf[T], error) {
	c, err := config.BuildDir(buildSerializerFromFactory(), configDir)
	if err != nil {
		return nil, err
	}

	conf := &configOf[T]{config: &server.Config{Dir: configDir}}
	if err := c.Load(name, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func buildSerializerFromFactory() config.Serializer {
	s := make(config.Serializer, len(filesFactory.items))
	for ext, item := range filesFactory.items {
		s.Add(item.Marshal, item.Unmarshal, ext)
	}
	return s
}

func init() {
	RegisterFileSerializer("json", json.Marshal, json.Unmarshal, ".json")
	RegisterFileSerializer("xml", xml.Marshal, xml.Unmarshal, ".xml")
	RegisterFileSerializer("yaml", yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
}
