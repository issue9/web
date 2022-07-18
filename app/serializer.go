// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"encoding/xml"
	"io/fs"

	"gopkg.in/yaml.v3"

	"github.com/issue9/web/internal/filesystem"
	"github.com/issue9/web/serializer"
)

var filesFactory = map[string]mimetype{}

// RegisterFileSerializer 注册 mimetype
//
// ext 为文件的扩展名，如果存在同名，则会覆盖。
// 同时 NewServerOf 也会根据此函数注册的扩展名进行解析。
func RegisterFileSerializer(m serializer.MarshalFunc, u serializer.UnmarshalFunc, ext ...string) {
	for _, e := range ext {
		filesFactory[e] = mimetype{m: m, u: u}
	}
}

func init() {
	RegisterFileSerializer(json.Marshal, json.Unmarshal, ".json")
	RegisterFileSerializer(xml.Marshal, xml.Unmarshal, ".xml")
	RegisterFileSerializer(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
}

func loadConfigOf[T any](fsys fs.FS, path string) (*configOf[T], error) {
	s := serializer.New(len(filesFactory))
	for name, ss := range filesFactory {
		if err := s.Add(ss.m, ss.u, name); err != nil {
			return nil, err
		}
	}

	f := filesystem.NewSerializer(s)
	conf := &configOf[T]{}
	if err := f.Load(fsys, path, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
