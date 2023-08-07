// SPDX-License-Identifier: MIT

// Package schema 将 ast 转换铖 openapi 的 schema 对象
package schema

import (
	"go/ast"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/issue9/web/cmd/web/internal/restdoc/utils"
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

type (
	Ref     = openapi3.SchemaRef
	OpenAPI = openapi3.T
)

func NewRef(ref string, v *openapi3.Schema) *Ref {
	return openapi3.NewSchemaRef(ref, v)
}

// NewOpenAPI 声明基本的 OpenAPI 对象
//
// 主要是对一些基本字段作为初始化。
func NewOpenAPI(ver string) *OpenAPI {
	c := openapi3.NewComponents()
	c.Schemas = make(openapi3.Schemas)
	c.Responses = make(openapi3.Responses)
	c.SecuritySchemes = make(openapi3.SecuritySchemes)

	return &openapi3.T{
		OpenAPI:    ver,
		Components: &c,
		Servers:    make(openapi3.Servers, 0, 5),
		Paths:      make(openapi3.Paths, 100),
		Security:   make(openapi3.SecurityRequirements, 0, 5),
		Tags:       make(openapi3.Tags, 0, 10),
	}
}

func addRefPrefix(ref *Ref) {
	if ref.Ref != "" && !strings.HasPrefix(ref.Ref, refPrefix) {
		ref.Ref = refPrefix + ref.Ref
	}
}

func parseTypeDoc(s *ast.TypeSpec) (title, desc, typ string, enums []any) {
	title, desc = parseComment(s.Comment, s.Doc)

	var lines []string
	if desc != "" {
		lines = strings.Split(desc, "\n")
	} else {
		lines = strings.Split(title, "\n")
	}

	for _, line := range lines {
		tag, suffix := utils.CutTag(line)
		switch tag {
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
