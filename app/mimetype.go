// SPDX-License-Identifier: MIT

package app

import (
	"strconv"

	"github.com/issue9/sliceutil"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/mimetype/form"
	"github.com/issue9/web/mimetype/html"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/mimetype/nop"
	"github.com/issue9/web/mimetype/xml"
	"github.com/issue9/web/server"
)

var mimetypesFactory = map[string]serializerItem{}

type serializerItem struct {
	marshal   web.MarshalFunc
	unmarshal web.UnmarshalFunc
}

type mimetypeConfig struct {
	// 编码名称
	//
	// 比如 application/xml 等
	Type string `json:"type" yaml:"type" xml:"type,attr"`

	// 返回错误代码是的 mimetype
	//
	// 比如正常情况下如果是 application/json，那么此值可以是 application/problem+json。
	// 如果为空，表示与 Type 相同。
	Problem string `json:"problem,omitempty" yaml:"problem,omitempty" xml:"problem,attr,omitempty"`

	// 实际采用的解码方法
	//
	// 由 [RegisterMimetype] 注册而来。默认可用为：
	//
	//  - xml
	//  - json
	//  - form
	//  - html
	//  - gob
	//  - nop  没有具体实现的方法，对于上传等需要自行处理的情况可以指定此值。
	Target string `json:"target" yaml:"target" xml:"target,attr"`
}

func (conf *configOf[T]) sanitizeMimetypes() *web.FieldError {
	indexes := sliceutil.Dup(conf.Mimetypes, func(i, j *mimetypeConfig) bool { return i.Type == j.Type })
	if len(indexes) > 0 {
		value := conf.Mimetypes[indexes[1]].Type
		err := web.NewFieldError("["+strconv.Itoa(indexes[1])+"].target", locales.DuplicateValue)
		err.Value = value
		return err
	}

	ms := make([]*server.Mimetype, 0, len(conf.Mimetypes))
	for index, item := range conf.Mimetypes {
		m, found := mimetypesFactory[item.Target]
		if !found {
			return web.NewFieldError("["+strconv.Itoa(index)+"].target", web.NewLocaleError("%s not found", item.Target))
		}

		ms = append(ms, &server.Mimetype{
			Marshal:   m.marshal,
			Unmarshal: m.unmarshal,
			Name:      item.Type,
			Problem:   item.Problem,
		})
	}
	conf.mimetypes = ms

	return nil
}

// RegisterMimetype 注册用于序列化用户提交数据的方法
//
// name 为名称，如果存在同名，则会覆盖。
func RegisterMimetype(m web.MarshalFunc, u web.UnmarshalFunc, name string) {
	mimetypesFactory[name] = serializerItem{marshal: m, unmarshal: u}
}

func init() {
	RegisterMimetype(json.Marshal, json.Unmarshal, "json")
	RegisterMimetype(xml.Marshal, xml.Unmarshal, "xml")
	RegisterMimetype(html.Marshal, html.Unmarshal, "html")
	RegisterMimetype(form.Marshal, form.Unmarshal, "form")
	RegisterMimetype(nop.Marshal, nop.Unmarshal, "nop")
}
