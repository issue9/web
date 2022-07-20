// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"encoding/xml"
	"io/fs"

	"github.com/issue9/localeutil"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/internal/serialization"
	"github.com/issue9/web/serializer"
	"github.com/issue9/web/serializer/form"
	"github.com/issue9/web/serializer/gob"
	"github.com/issue9/web/serializer/html"
	"github.com/issue9/web/serializer/protobuf"
)

var (
	mimetypesFactory = map[string]serialization.Item{}
	filesFactory     = map[string]serialization.Item{}
)

type mimetypeConfig struct {
	// 编码名称
	//
	// 比如 application/xml 等
	Encoding string `json:"encoding" yaml:"encoding" xml:"encoding,attr"`

	// 实际采用的解码方法
	//
	// 由 RegisterMimetype 注册而来。可通过 RegisterMimetype 注册新的格式，默认可用为：
	//
	//  - xml
	//  - json
	//  - protobuf
	//  - gob
	//  - form
	//  - html
	//  - nil
	Target string `json:"target" yaml:"target" xml:"target,attr"`
}

func (conf *configOf[T]) buildMimetypes(mt serializer.Serializer) *ConfigError {
	for _, item := range conf.Mimetypes {
		m, found := mimetypesFactory[item.Target]
		if !found {
			return &ConfigError{Field: item.Target, Message: localeutil.Error("%s not found", item.Target)}
		}

		if err := mt.Add(m.Marshal, m.Unmarshal, item.Encoding); err != nil {
			return &ConfigError{Field: item.Target, Message: err}
		}
	}

	return nil
}

// RegisterMimetype 注册 mimetype
//
// name 为名称，如果存在同名，则会覆盖。
func RegisterMimetype(m serializer.MarshalFunc, u serializer.UnmarshalFunc, name string) {
	mimetypesFactory[name] = serialization.Item{Marshal: m, Unmarshal: u}
}

// RegisterFileSerializer 注册 mimetype
//
// ext 为文件的扩展名，如果存在同名，则会覆盖。
func RegisterFileSerializer(m serializer.MarshalFunc, u serializer.UnmarshalFunc, ext ...string) {
	for _, e := range ext {
		filesFactory[e] = serialization.Item{Marshal: m, Unmarshal: u}
	}
}

func loadConfigOf[T any](fsys fs.FS, path string) (*configOf[T], error) {
	f := serialization.NewFS(len(filesFactory))
	s := f.Serializer()
	for name, ss := range filesFactory {
		if err := s.Add(ss.Marshal, ss.Unmarshal, name); err != nil {
			return nil, err
		}
	}

	conf := &configOf[T]{}
	if err := f.Load(fsys, path, conf); err != nil {
		return nil, err
	}

	if err := conf.sanitize(); err != nil {
		err.Path = path
		return nil, err
	}

	return conf, nil
}

func init() {
	RegisterMimetype(json.Marshal, json.Unmarshal, "json")
	RegisterMimetype(xml.Marshal, xml.Unmarshal, "xml")
	RegisterMimetype(nil, nil, "nil")
	RegisterMimetype(gob.Marshal, gob.Unmarshal, "gob")
	RegisterMimetype(html.Marshal, html.Unmarshal, "html")
	RegisterMimetype(form.Marshal, form.Unmarshal, "form")
	RegisterMimetype(protobuf.Marshal, protobuf.Unmarshal, "protobuf")

	RegisterFileSerializer(json.Marshal, json.Unmarshal, ".json")
	RegisterFileSerializer(xml.Marshal, xml.Unmarshal, ".xml")
	RegisterFileSerializer(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
}
