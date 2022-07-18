// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"encoding/xml"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/serializer"
	"github.com/issue9/web/serializer/form"
	"github.com/issue9/web/serializer/gob"
	"github.com/issue9/web/serializer/html"
	"github.com/issue9/web/serializer/protobuf"
)

var mimetypesFactory = map[string]mimetype{}

type mimetype struct {
	m serializer.MarshalFunc
	u serializer.UnmarshalFunc
}

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

func (conf *configOf[T]) buildMimetypes(mt *serializer.Serializer) *ConfigError {
	for _, item := range conf.Mimetypes {
		m, found := mimetypesFactory[item.Target]
		if !found {
			return &ConfigError{Field: item.Target, Message: localeutil.Error("%s not found", item.Target)}
		}

		if err := mt.Add(m.m, m.u, item.Encoding); err != nil {
			return &ConfigError{Field: item.Target, Message: err}
		}
	}

	return nil
}

// RegisterMimetype 注册 mimetype
//
// name 为名称，如果存在同名，则会覆盖。
func RegisterMimetype(m serializer.MarshalFunc, u serializer.UnmarshalFunc, name string) {
	mimetypesFactory[name] = mimetype{m: m, u: u}
}

func init() {
	RegisterMimetype(json.Marshal, json.Unmarshal, "json")
	RegisterMimetype(xml.Marshal, xml.Unmarshal, "xml")
	RegisterMimetype(nil, nil, "nil")
	RegisterMimetype(gob.Marshal, gob.Unmarshal, "gob")
	RegisterMimetype(html.Marshal, html.Unmarshal, "html")
	RegisterMimetype(form.Marshal, form.Unmarshal, "form")
	RegisterMimetype(protobuf.Marshal, protobuf.Unmarshal, "protobuf")
}
