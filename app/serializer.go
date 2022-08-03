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
	"github.com/issue9/web/serializer/html"
	"github.com/issue9/web/server"
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

	// 对应 Problem 类型的编码名称
	//
	// 如果为空，表示与 encoding 相同，根据 [RFC7807] 最好是不相同，
	// 比如 application/json 对应 application/problem+json。
	//
	// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
	Problem string `json:"problem,omitempty" yaml:"problem,omitempty" xml:"problem,attr,omitempty"`

	// 实际采用的解码方法
	//
	// 由 [RegisterMimetype] 注册而来。默认可用为：
	//
	//  - xml
	//  - json
	//  - form
	//  - html
	//  - nil
	Target string `json:"target" yaml:"target" xml:"target,attr"`
}

func (conf *configOf[T]) buildMimetypes(srv *server.Server) *ConfigError {
	problems := make(map[string]string, len(conf.Mimetypes))

	for _, item := range conf.Mimetypes {
		m, found := mimetypesFactory[item.Target]
		if !found {
			return &ConfigError{Field: item.Target, Message: localeutil.Error("%s not found", item.Target)}
		}

		if err := srv.Mimetypes().Add(m.Marshal, m.Unmarshal, item.Encoding); err != nil {
			return &ConfigError{Field: item.Target, Message: err}
		}

		if item.Problem != "" {
			problems[item.Encoding] = item.Problem
		}
	}

	for k, v := range problems {
		srv.Problems().AddMimetype(k, v)
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
	RegisterMimetype(html.Marshal, html.Unmarshal, "html")
	RegisterMimetype(form.Marshal, form.Unmarshal, "form")

	RegisterFileSerializer(json.Marshal, json.Unmarshal, ".json")
	RegisterFileSerializer(xml.Marshal, xml.Unmarshal, ".xml")
	RegisterFileSerializer(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
}
