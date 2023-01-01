// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"encoding/xml"
	"io/fs"
	"path/filepath"
	"strconv"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/serializer/form"
	"github.com/issue9/web/serializer/html"
	"github.com/issue9/web/server"
)

var (
	mimetypesFactory = map[string]serializerItem{}
	filesFactory     = map[string]fileSerializerItem{}
)

type serializerItem struct {
	marshal   server.MarshalFunc
	unmarshal server.UnmarshalFunc
}

type fileSerializerItem struct {
	marshal   server.MarshalFileFunc
	unmarshal server.UnmarshalFileFunc
}

type mimetypeConfig struct {
	// 编码名称
	//
	// 比如 application/xml 等
	Type string `json:"type" yaml:"type" xml:"type,attr"`

	// 对应 [server.Problem] 类型的编码名称
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

type mimetype struct {
	m       server.MarshalFunc
	u       server.UnmarshalFunc
	name    string
	problem string
}

func (conf *configOf[T]) sanitizeMimetypes() *ConfigError {
	dup := sliceutil.Dup(conf.Mimetypes, func(i, j *mimetypeConfig) bool { return i.Type == j.Type })
	if len(dup) > 0 {
		value := conf.Mimetypes[dup[1]].Type
		return &ConfigError{Field: "[" + strconv.Itoa(dup[1]) + "].target", Message: localeutil.Phrase("duplicate value"), Value: value}
	}

	ms := make([]mimetype, 0, len(conf.Mimetypes))
	for index, item := range conf.Mimetypes {
		m, found := mimetypesFactory[item.Target]
		if !found {
			return &ConfigError{Field: "[" + strconv.Itoa(index) + "].target", Message: localeutil.Phrase("%s not found", item.Target)}
		}

		ms = append(ms, mimetype{m: m.marshal, u: m.unmarshal, name: item.Type, problem: item.Problem})
	}
	conf.mimetypes = ms

	return nil
}

func (conf *configOf[T]) sanitizeFiles() *ConfigError {
	conf.files = make(map[string]fileSerializerItem, len(conf.Files))
	for i, name := range conf.Files {
		s, found := filesFactory[name]
		if !found {
			return &ConfigError{Field: "[" + strconv.Itoa(i) + "]", Message: localeutil.Phrase("not found serialization function for %s", name)}
		}
		conf.files[name] = s // conf.Files 可以保证 conf.files 唯一性
	}
	return nil
}

// RegisterMimetype 注册用于序列化用户提交数据的方法
//
// name 为名称，如果存在同名，则会覆盖。
func RegisterMimetype(m server.MarshalFunc, u server.UnmarshalFunc, name string) {
	mimetypesFactory[name] = serializerItem{marshal: m, unmarshal: u}
}

// RegisterFileSerializer 注册用于文件序列化的方法
//
// ext 为文件的扩展名，如果存在同名，则会覆盖。
func RegisterFileSerializer(m server.MarshalFileFunc, u server.UnmarshalFileFunc, ext ...string) {
	for _, e := range ext {
		filesFactory[e] = fileSerializerItem{marshal: m, unmarshal: u}
	}
}

func loadConfigOf[T any](fsys fs.FS, path string) (*configOf[T], error) {
	ext := filepath.Ext(path)
	var item fileSerializerItem
	for name, ss := range filesFactory {
		if name == ext {
			item = ss
			break
		}
	}

	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}

	conf := &configOf[T]{}
	if err := item.unmarshal(data, conf); err != nil {
		return nil, err
	}

	if err := conf.sanitize(); err != nil {
		err.Path = path
		return nil, err
	}

	return conf, nil
}

func init() {
	RegisterMimetype(server.MarshalJSON, json.Unmarshal, "json")
	RegisterMimetype(server.MarshalXML, xml.Unmarshal, "xml")
	RegisterMimetype(nil, nil, "nil")
	RegisterMimetype(html.Marshal, html.Unmarshal, "html")
	RegisterMimetype(form.Marshal, form.Unmarshal, "form")

	RegisterFileSerializer(json.Marshal, json.Unmarshal, ".json")
	RegisterFileSerializer(xml.Marshal, xml.Unmarshal, ".xml")
	RegisterFileSerializer(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
}
