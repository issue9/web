// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package openapi 对 openapi 的再处理
package openapi

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/web"
	"gopkg.in/yaml.v3"
)

// OpenAPI 协程安全的 OpenAPI 对象
type OpenAPI struct {
	doc          *openapi3.T
	pathLocker   sync.Mutex
	schemaLocker sync.RWMutex
}

func New(ver string) *OpenAPI {
	c := openapi3.NewComponents()
	c.Schemas = make(openapi3.Schemas)
	c.Responses = make(openapi3.ResponseBodies)
	c.SecuritySchemes = make(openapi3.SecuritySchemes)

	return &OpenAPI{doc: &openapi3.T{
		OpenAPI:    ver,
		Components: &c,
		Servers:    make(openapi3.Servers, 0, 5),
		Paths:      openapi3.NewPathsWithCapacity(100),
		Security:   make(openapi3.SecurityRequirements, 0, 5),
		Tags:       make(openapi3.Tags, 0, 10),
	}}
}

func (doc *OpenAPI) Doc() *openapi3.T { return doc.doc }

// SaveAs 保存为 yaml 或 json 文件
//
// 根据后缀名名确定保存的文件类型，目前仅支持 json 和 yaml。
func (doc *OpenAPI) SaveAs(path string) error {
	var m func(any) ([]byte, error)
	switch filepath.Ext(path) {
	case ".yaml", ".yml": // BUG: 依赖的 openapi3.Paths 不支持输出 yaml?
		m = yaml.Marshal
	case ".json":
		m = func(v any) ([]byte, error) { return json.MarshalIndent(v, "", "\t") }
	default:
		return web.NewLocaleError("only support yaml and json")
	}

	data, err := m(doc.Doc())
	if err == nil {
		err = os.WriteFile(path, data, os.ModePerm)
	}
	return err
}

func (doc *OpenAPI) Merge(d *openapi3.T) {
	doc.Doc().Servers = append(doc.Doc().Servers, d.Servers...)
	doc.Doc().Security = append(doc.Doc().Security, d.Security...)
	doc.Doc().Tags = append(doc.Doc().Tags, d.Tags...)
	if d.Paths != nil && d.Paths.Len() > 0 {
		for k, v := range d.Paths.Map() {
			doc.Doc().Paths.Set(k, v)
		}
	}
	if d.Components != nil {
		maps.Copy(doc.Doc().Components.Schemas, d.Components.Schemas)
		maps.Copy(doc.Doc().Components.Parameters, d.Components.Parameters)
		maps.Copy(doc.Doc().Components.Headers, d.Components.Headers)
		maps.Copy(doc.Doc().Components.RequestBodies, d.Components.RequestBodies)
		maps.Copy(doc.Doc().Components.Responses, d.Components.Responses)
		maps.Copy(doc.Doc().Components.SecuritySchemes, d.Components.SecuritySchemes)
		maps.Copy(doc.Doc().Components.Examples, d.Components.Examples)
		maps.Copy(doc.Doc().Components.Links, d.Components.Links)
		maps.Copy(doc.Doc().Components.Callbacks, d.Components.Callbacks)
	}
}
