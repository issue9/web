// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"encoding/xml"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/serialization"
	"github.com/issue9/web/serialization/form"
	"github.com/issue9/web/serialization/gob"
	"github.com/issue9/web/serialization/html"
	"github.com/issue9/web/serialization/protobuf"
)

var mimetypesFactory = map[string]mimetype{}

type mimetype struct {
	m serialization.MarshalFunc
	u serialization.UnmarshalFunc
}

func (conf *configOf[T]) buildMimetypes() (*serialization.Mimetypes, *ConfigError) {
	mt := serialization.NewMimetypes(len(conf.Mimetypes))

	for _, name := range conf.Mimetypes {
		m, found := mimetypesFactory[name]
		if !found {
			return nil, &ConfigError{Field: name, Message: localeutil.Error("%s not found", name)}
		}

		if err := mt.Add(m.m, m.u, name); err != nil {
			return nil, &ConfigError{Field: name, Message: err}
		}
	}

	return mt, nil
}

// RegisterMimetype 注册 mimetype
//
// name 为缓存的名称，如果存在同名，则会覆盖。
func RegisterMimetype(m serialization.MarshalFunc, u serialization.UnmarshalFunc, name ...string) {
	if len(name) == 0 {
		panic("参数 name 不能为空")
	}

	for _, n := range name {
		mimetypesFactory[n] = mimetype{m: m, u: u}
	}
}

func init() {
	RegisterMimetype(json.Marshal, json.Unmarshal, "application/json")
	RegisterMimetype(xml.Marshal, xml.Unmarshal, "application/xml", "text/xml")
	RegisterMimetype(gob.Marshal, gob.Unmarshal, gob.Mimetype)
	RegisterMimetype(html.Marshal, html.Unmarshal, html.Mimetype)
	RegisterMimetype(form.Marshal, form.Unmarshal, form.Mimetype)
	RegisterMimetype(protobuf.Marshal, protobuf.Unmarshal, protobuf.Mimetype)
}
