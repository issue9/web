// SPDX-License-Identifier: MIT

// Package schema 将 ast 转换铖 openapi 的 schema 对象
package schema

import (
	"go/ast"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/issue9/web/cmd/web/restdoc/utils"
)

const refPrefix = "#/components/schemas/"

var refReplacer = strings.NewReplacer(
	"/", ".",
	"[", "-",
	"]", "-",
	",", "--",
	" ", "",
	"\t", "",
)

type Ref = openapi3.SchemaRef

func addRefPrefix(ref string) string {
	if !strings.HasPrefix(ref, refPrefix) {
		ref = refPrefix + ref
	}
	return ref
}

func NewRef(ref string, v *openapi3.Schema) *Ref {
	return openapi3.NewSchemaRef(ref, v)
}

func parseTypeDoc(s *ast.TypeSpec) (title, desc, typ string, enums []any) {
	title, desc = parseComment(s.Comment, s.Doc)

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

	return title, desc, typ, enums
}

func parseComment(comments, doc *ast.CommentGroup) (title, desc string) {
	if doc == nil {
		doc = comments
	}
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

func getPrimitiveType(name string, isArray bool) (*Ref, bool) {
	switch name { // 基本类型
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return array(NewRef("", openapi3.NewIntegerSchema()), isArray), true
	case "float32", "float64":
		return array(NewRef("", openapi3.NewFloat64Schema()), isArray), true
	case "bool":
		return array(NewRef("", openapi3.NewBoolSchema()), isArray), true
	case "string":
		return array(NewRef("", openapi3.NewStringSchema()), isArray), true
	case "map":
		return array(NewRef("", openapi3.NewObjectSchema()), isArray), true
	case "{}":
		return nil, true

	// 以下是对一些内置类型的特殊处理
	case "time.Time":
		return array(NewRef("", openapi3.NewDateTimeSchema()), isArray), true
	case "time.Duration":
		return array(NewRef("", openapi3.NewStringSchema()), isArray), true
	default:
		return nil, false
	}
}

// 根据 isArray 将 ref 包装成相应的对象
func array(ref *Ref, isArray bool) *Ref {
	if !isArray {
		return ref
	}

	s := openapi3.NewArraySchema()
	s.Items = ref
	return NewRef("", s)
}

// 将从 components/schemas 中获取的对象进行二次包装
func wrap(ref *Ref, title, desc string, xml *openapi3.XML, nullable bool) *Ref {
	if ref == nil {
		return ref
	}

	if ref.Ref == "" { // 非引用模式，表示该值仅调用方使用，直接修改值。
		if desc != "" {
			ref.Value.Description = desc
		}
		if title != "" {
			ref.Value.Title = title
		}

		if ref.Value.XML != xml {
			ref.Value.XML = xml
		}

		if ref.Value.Nullable != nullable {
			ref.Value.Nullable = nullable
		}

		return ref
	}

	if ref.Value.Nullable != nullable ||
		ref.Value.XML != xml ||
		(desc != "" && ref.Value.Description != desc) ||
		(title != "" && ref.Value.Title != title) {
		s := openapi3.NewSchema()
		s.AllOf = openapi3.SchemaRefs{ref}
		s.Nullable = nullable
		s.XML = xml
		if desc != "" {
			s.Description = desc
		}
		if title != "" {
			s.Title = title
		}
		ref = NewRef("", s)
	}
	return ref
}
