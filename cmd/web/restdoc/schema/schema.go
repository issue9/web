// SPDX-License-Identifier: MIT

// Package schema 将 ast 转换铖 openapi 的 schema 对象
package schema

import (
	"go/ast"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/query/v3"

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
}

func New(l *logger.Logger) *Schema { return &Schema{pkg: pkg.New(l)} }

// Packages 返回关联的 [pkg.Packages]
func (s *Schema) Packages() *pkg.Packages { return s.pkg }

// 从 title 和 desc 中获取类型名称和枚举信息
func parseTypeDoc(title, desc string) (typ string, enums []any) {
	// TODO enums 如果为空，则从代码中获取？
	var lines []string
	if desc != "" {
		lines = strings.Split(desc, "\n")
	} else {
		lines = []string{title}
	}

	for _, line := range lines {
		switch tag, suffix := utils.CutTag(line); tag {
		case "@enum", "@enums": // @enum e1 e2 e3
			for _, word := range strings.Fields(suffix) {
				enums = append(enums, word)
			}
		case "@type": // @type string
			typ = suffix
		}
	}

	return typ, enums
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

func buildBasicType(name string) (*openapi3.SchemaRef, bool) {
	switch name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return openapi.NewSchemaRef("", openapi3.NewIntegerSchema()), true
	case "float32", "float64":
		return openapi.NewSchemaRef("", openapi3.NewFloat64Schema()), true
	case "bool":
		return openapi.NewSchemaRef("", openapi3.NewBoolSchema()), true
	case "string":
		return openapi.NewSchemaRef("", openapi3.NewStringSchema()), true
	case "map":
		return openapi.NewSchemaRef("", openapi3.NewObjectSchema()), true
	case "{}":
		return nil, true
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
