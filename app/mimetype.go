// SPDX-License-Identifier: MIT

package app

import (
	"strconv"

	"github.com/issue9/config"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/serializer/form"
	"github.com/issue9/web/serializer/html"
	"github.com/issue9/web/serializer/json"
	"github.com/issue9/web/serializer/xml"
	"github.com/issue9/web/server"
)

var mimetypesFactory = map[string]serializerItem{}

type serializerItem struct {
	marshal   server.MarshalFunc
	unmarshal server.UnmarshalFunc
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
	//  - nil  未实际指定序列化方法，最终需要用户自行处理，比如返回文件上传等。
	Target string `json:"target" yaml:"target" xml:"target,attr"`
}

func (conf *configOf[T]) sanitizeMimetypes() *config.FieldError {
	dup := sliceutil.Dup(conf.Mimetypes, func(i, j *mimetypeConfig) bool { return i.Type == j.Type })
	if len(dup) > 0 {
		value := conf.Mimetypes[dup[1]].Type
		err := config.NewFieldError("["+strconv.Itoa(dup[1])+"].target", errs.NewLocaleError("duplicate value"))
		err.Value = value
		return err
	}

	ms := make([]*server.Mimetype, 0, len(conf.Mimetypes))
	for index, item := range conf.Mimetypes {
		m, found := mimetypesFactory[item.Target]
		if !found {
			return config.NewFieldError("["+strconv.Itoa(index)+"].target", errs.NewLocaleError("%s not found", item.Target))
		}

		ms = append(ms, &server.Mimetype{
			Marshal:     m.marshal,
			Unmarshal:   m.unmarshal,
			Type:        item.Type,
			ProblemType: item.Problem,
		})
	}
	conf.mimetypes = ms

	return nil
}

// RegisterMimetype 注册用于序列化用户提交数据的方法
//
// name 为名称，如果存在同名，则会覆盖。
func RegisterMimetype(m server.MarshalFunc, u server.UnmarshalFunc, name string) {
	mimetypesFactory[name] = serializerItem{marshal: m, unmarshal: u}
}

func init() {
	RegisterMimetype(json.Marshal, json.Unmarshal, "json")
	RegisterMimetype(xml.Marshal, xml.Unmarshal, "xml")
	RegisterMimetype(nil, nil, "nil")
	RegisterMimetype(html.Marshal, html.Unmarshal, "html")
	RegisterMimetype(form.Marshal, form.Unmarshal, "form")
}
