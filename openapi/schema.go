// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"reflect"
	"strings"
	"time"

	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/issue9/web"
)

// OpenAPISchema 自定义某个类型在 openapi 文档中的类型
type OpenAPISchema interface {
	// OpenAPISchema 修改当前类型的 [Schema] 表示形式
	OpenAPISchema(s *Schema)
}

var openAPISchemaType = reflect.TypeFor[OpenAPISchema]()

type Parameter struct {
	Ref *Ref

	Name        string
	Deprecated  bool
	Required    bool
	Description web.LocaleStringer // 当前参数的描述信息

	// InHeader 模式下此值无效
	Schema *Schema
}

type parameterRenderer struct {
	Name        string                    `json:"name" yaml:"name"`
	In          string                    `json:"in" yaml:"in"`
	Required    bool                      `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated  bool                      `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Schema      *renderer[schemaRenderer] `json:"schema,omitempty" yaml:"schema,omitempty"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
}

type headerRenderer struct {
	Required   bool `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

type Schema struct {
	Ref *Ref

	XML                  *XML
	ExternalDocs         *ExternalDocs
	Title                web.LocaleStringer
	Description          web.LocaleStringer
	Type                 string
	AllOf                []*Schema
	OneOf                []*Schema
	AnyOf                []*Schema
	Format               string
	Items                *Schema
	Properties           map[string]*Schema
	AdditionalProperties *Schema
	Required             []string
	Minimum              int
	Maximum              int
	Enum                 []any
	Default              any
}

type properties = orderedmap.OrderedMap[string, *renderer[schemaRenderer]]

type schemaRenderer struct {
	Type                 string                      `json:"type,omitempty" yaml:"type,omitempty"` // AnyOf 等不为空，此值可为空
	XML                  *XML                        `json:"xml,omitempty" yaml:"xml,omitempty"`
	ExternalDocs         *externalDocsRenderer       `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Title                string                      `json:"title,omitempty" yaml:"title,omitempty"`
	Description          string                      `json:"description,omitempty" yaml:"description,omitempty"`
	AllOf                []*renderer[schemaRenderer] `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf                []*renderer[schemaRenderer] `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                []*renderer[schemaRenderer] `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Format               string                      `json:"format,omitempty" yaml:"format,omitempty"`
	Items                *renderer[schemaRenderer]   `json:"items,omitempty" yaml:"items,omitempty"`
	Properties           *properties                 `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties *renderer[schemaRenderer]   `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Required             []string                    `json:"required,omitempty" yaml:"required,omitempty"`
	Minimum              int                         `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum              int                         `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	Enum                 []any                       `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default              any                         `json:"default,omitempty" yaml:"default,omitempty"`
}

func (d *Document) newSchema(v any) *Schema { return newSchema(d, v, nil, nil) }

// NewSchema 根据 v 生成 [Schema] 对象
//
// 如果 v 不是空值，那么 v 也将同时作为默认值出现在 [Schema] 中。
func NewSchema(v any, title, desc web.LocaleStringer) *Schema {
	return newSchema(nil, v, title, desc)
}

func AnyOfSchema(title, desc web.LocaleStringer, v ...any) *Schema {
	return xOfSchema(0, title, desc, v...)
}

func OneOfSchema(title, desc web.LocaleStringer, v ...any) *Schema {
	return xOfSchema(1, title, desc, v...)
}

func AllOfSchema(title, desc web.LocaleStringer, v ...any) *Schema {
	return xOfSchema(2, title, desc, v...)
}

// - 0 AnyOf
// - 1 OneOf
// - 2 AllOf
func xOfSchema(typ int, title, desc web.LocaleStringer, v ...any) *Schema {
	if len(v) == 0 {
		panic("参数 v 必不可少")
	}

	of := make([]*Schema, 0, len(v))
	for _, vv := range v {
		of = append(of, NewSchema(vv, nil, nil))
	}

	s := &Schema{
		Title:       title,
		Description: desc,
	}

	switch typ {
	case 0:
		s.AnyOf = of
	case 1:
		s.OneOf = of
	case 2:
		s.AllOf = of
	default:
		panic("无效的参数 typ")
	}

	return s
}

func newSchema(d *Document, v any, title, desc web.LocaleStringer) *Schema {
	if v == nil {
		return nil
	}

	s := &Schema{
		Title:       title,
		Description: desc,
	}

	rv := reflect.ValueOf(v)
	if !rv.IsZero() {
		s.Default = v
	}
	schemaFromType(d, rv.Type(), true, "", s)

	return s
}

var timeType = reflect.TypeFor[time.Time]()

// d 仅用于查找其关联的 components/schemas 中是否存在相同名称的对象，如果存在则直接生成引用对象。
//
// desc 表示类型 t 的 Description 属性
// rootName 根结构体的名称，主要是为了解决子元素又引用了根元素的类型引起的循环引用。
func schemaFromType(d *Document, t reflect.Type, isRoot bool, rootName string, s *Schema) {
	if t.Implements(openAPISchemaType) {
		if t.Kind() == reflect.Pointer { // 值类型的指针符合 t.Implements，但是无法使用 reflect.New(t).Elem 获得一个有效的值。
			t = t.Elem()
		}
		reflect.New(t).Interface().(OpenAPISchema).OpenAPISchema(s)
		return
	}

	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		s.Type = TypeString
	case reflect.Bool:
		s.Type = TypeBoolean
	case reflect.Float32, reflect.Float64:
		s.Type = TypeNumber
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s.Type = TypeInteger
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s.Type = TypeInteger
		s.Minimum = 0
	case reflect.Array, reflect.Slice:
		s.Type = TypeArray
		s.Items = &Schema{}

		if s.XML != nil {
			if index := strings.IndexByte(s.XML.Name, '>'); index > 0 {
				s.Items.XML = &XML{Name: s.XML.Name[index+1:]}
				s.XML.Name = s.XML.Name[:index]
				s.XML.Wrapped = true
			}
		}
		schemaFromType(d, t.Elem(), false, rootName, s.Items)
	case reflect.Map:
		s.Type = TypeObject
		s.AdditionalProperties = &Schema{}
		schemaFromType(d, t.Elem(), false, rootName, s.AdditionalProperties)
	case reflect.Struct:
		if t == timeType { // 对时间作特殊处理
			s.Type = TypeString
			s.Format = FormatDateTime
			return
		}
		schemaFromObjectType(d, t, isRoot, rootName, s)
	}
}

func schemaFromObjectType(d *Document, t reflect.Type, isRoot bool, rootName string, s *Schema) {
	typeName := getTypeName(t)

	if d != nil {
		if _, found := d.components.schemas[typeName]; found { // 已经存在于 components
			s.Ref = &Ref{Ref: typeName}
			return
		}
	}

	s.Ref = &Ref{Ref: typeName}
	if isRoot {
		rootName = typeName // isRoot == true 时，rootName 必然为空
	} else if typeName == rootName { // 在字段中引用了根对象
		return
	}

	s.Type = TypeObject
	s.Properties = make(map[string]*Schema, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if f.Anonymous {
			schemaFromType(d, f.Type, isRoot, rootName, s)
			continue
		}

		if k := f.Type.Kind(); !f.IsExported() || k == reflect.Chan || k == reflect.Func || k == reflect.Complex64 || k == reflect.Complex128 {
			continue
		}

		var itemDesc web.LocaleStringer
		name := f.Name
		var xml *XML
		if f.Tag != "" {
			tag, omitempty, _ := getTagName(f, "json")
			if tag == "-" {
				continue
			} else if tag != "" {
				name = tag
			}

			if !omitempty {
				s.Required = append(s.Required, name)
			}

			if xmlName, _, attr := getTagName(f, "xml"); xmlName != "" && xmlName != name {
				xml = &XML{Name: xmlName, Attribute: attr}
			}

			comment := f.Tag.Get(CommentTag)
			if comment != "" {
				itemDesc = web.Phrase(comment)
			}

			if tt := f.Tag.Get("openapi"); tt != "" {
				if tt != "-" {
					tags := strings.Split(tt, ",")

					format := ""
					if len(tags) > 1 {
						format = tags[1]
					}

					s.Properties[name] = &Schema{
						Description: itemDesc,
						Type:        tags[0],
						Format:      format,
						XML:         xml,
					}
				}
				continue
			}
		} // end f.Tag

		item := &Schema{
			Description: itemDesc,
			XML:         xml,
		}
		schemaFromType(d, t.Field(i).Type, false, rootName, item)
		if item.Type == "" {
			continue
		}

		s.Properties[name] = item
	}
}

func (s *Schema) isBasicType() bool {
	switch s.Type {
	case TypeObject:
		return false
	case TypeArray:
		return s.Items.isBasicType()
	default:
		return true
	}
}
