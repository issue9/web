// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package schema 将 ast 转换铖 openapi 的 schema 对象
package schema

import (
	"errors"
	"go/ast"
	"go/types"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/query/v3"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/enum"
	"github.com/issue9/web/cmd/web/restdoc/logger"
	"github.com/issue9/web/cmd/web/restdoc/openapi"
	"github.com/issue9/web/cmd/web/restdoc/pkg"
	"github.com/issue9/web/cmd/web/restdoc/utils"
)

var refReplacer = strings.NewReplacer(
	"/", ".",
	"[", "-",
	"]", "-",
	",", "--",
	" ", "",
	"\t", "",
)

// Schema 管理 Schema 的查询
type Schema struct {
	pkg *pkg.Packages

	structs  map[string]*openapi3.SchemaRef
	structsM sync.Mutex
}

func New(l *logger.Logger) *Schema {
	return &Schema{
		pkg:     pkg.New(l),
		structs: make(map[string]*openapi3.SchemaRef, 10),
	}
}

// 判断 t 是否已经被申明，如果已经申明则返回其关联的 [openapi3.SchemaRef] 对象，否则返回空对象并标记该对象已声明。
func (s *Schema) getStruct(ref string, t types.Type) *openapi3.SchemaRef {
	s.structsM.Lock()
	defer s.structsM.Unlock()

	if r, found := s.structs[t.String()]; found {
		return r
	}

	if ref != "" {
		s.structs[t.String()] = openapi.NewSchemaRef(ref, nil)
	}

	return nil
}

// Packages 返回关联的 [pkg.Packages]
func (s *Schema) Packages() *pkg.Packages { return s.pkg }

// 从 title 和 desc 中获取类型名称和枚举信息
func parseTypeDoc(obj *types.TypeName, title, desc string) (typ string, enums []string, err error) {
	var lines []string
	if desc != "" {
		lines = strings.Split(desc, "\n")
	} else {
		lines = []string{title}
	}

	for _, line := range lines {
		switch tag, suffix := utils.CutTag(line); tag {
		case "@enum", "@enums": // @enum e1 e2 e3
			enums = strings.Fields(suffix)
			if len(enums) == 0 { // 如果为空，则从代码中提取
				enums, err = getEnums(obj.Pkg(), obj.Name())
				if err != nil {
					return "", nil, err
				}
			}
		case "@type": // @type string
			typ = suffix
		}
	}

	return typ, enums, nil
}

func getEnums(pkg *types.Package, t string) ([]string, error) {
	vals, err := enum.GetValue(pkg, t)
	if errors.Is(err, enum.ErrNotAllowedType) {
		return nil, web.NewLocaleError("@enum can not be empty")
	} else if err != nil {
		return nil, err
	}

	hasPrefix := true
	for _, v := range vals {
		if hasPrefix = strings.HasPrefix(v, t); !hasPrefix {
			break
		}
	}

	for i, v := range vals {
		if hasPrefix {
			v = strings.TrimPrefix(v, t)
		}
		rs := []rune(v)
		rs[0] = unicode.ToLower(rs[0])
		vals[i] = string(rs)
	}

	return vals, nil
}

func parseComment(doc *ast.CommentGroup) (title, desc string) {
	if doc == nil {
		return
	}

	if len(doc.List) == 1 {
		title = doc.Text()
		return title[:len(title)-1], ""
	}

	desc = doc.Text()
	if index := strings.IndexByte(desc, '\n'); index >= 0 {
		title = desc[:index]
	}

	return
}

func buildBasicType(t types.Type) (*openapi3.SchemaRef, bool) {
	if t == nil {
		return nil, true
	}
	if _, ok := t.(*types.Interface); ok {
		return openapi.NewSchemaRef("", openapi3.NewObjectSchema()), true
	}

	switch t.String() {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return openapi.NewSchemaRef("", openapi3.NewIntegerSchema()), true
	case "float32", "float64":
		return openapi.NewSchemaRef("", openapi3.NewFloat64Schema()), true
	case "bool":
		return openapi.NewSchemaRef("", openapi3.NewBoolSchema()), true
	case "string":
		return openapi.NewSchemaRef("", openapi3.NewStringSchema()), true
	case "time.Duration":
		return openapi.NewSchemaRef("", openapi3.NewStringSchema()), true
	case "time.Time":
		return openapi.NewSchemaRef("", openapi3.NewDateTimeSchema()), true
	default:
		return nil, false
	}
}

func parseTag(fieldName, tagValue, tagName string) (name string, nullable bool, xml *openapi3.XML) {
	if tagValue == "" {
		return fieldName, false, nil
	}

	structTag := reflect.StructTag(strings.Trim(tagValue, "`"))
	tag := structTag.Get(tagName)
	if tag == "-" { // 忽略此字段
		return "-", false, nil
	}

	if tag != "" {
		words := strings.Split(tag, ",")
		name = words[0]
		if len(words) > 1 && words[1] == "omitempty" {
			nullable = true
		}
	}

	if tagName != query.Tag { // 非查询参数对象，需要处理 XML 的特殊情况
		tag := structTag.Get("xml")
		if tag != "" && tag != "-" {
			words := strings.Split(tag, ",")
			switch len(words) {
			case 1:
				if index := strings.IndexByte(words[0], '>'); index > 0 {
					xml = &openapi3.XML{Wrapped: true, Name: words[0][index+1:]}
				}
			case 2:
				wrapIndex := strings.IndexByte(words[0], '>')
				attr := words[1] == "attr"
				if wrapIndex > 0 || attr {
					xml = &openapi3.XML{Wrapped: wrapIndex > 0, Attribute: attr}
					if xml.Wrapped {
						xml.Name = words[0][wrapIndex+1:]
					}
				}
			}
		}
	}

	return
}
