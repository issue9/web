// SPDX-License-Identifier: MIT

// Package openapi 对 openapi 的再处理
package openapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/web"
	"gopkg.in/yaml.v3"
)

// OpenAPI 协程安全的 OpenAPI 对象
type OpenAPI struct {
	doc *openapi3.T

	pathLocker   sync.Mutex
	schemaLocker sync.RWMutex
}

func New(ver string) *OpenAPI {
	c := openapi3.NewComponents()
	c.Schemas = make(openapi3.Schemas)
	c.Responses = make(openapi3.Responses)
	c.SecuritySchemes = make(openapi3.SecuritySchemes)

	t := &openapi3.T{
		OpenAPI:    ver,
		Components: &c,
		Servers:    make(openapi3.Servers, 0, 5),
		Paths:      make(openapi3.Paths, 100),
		Security:   make(openapi3.SecurityRequirements, 0, 5),
		Tags:       make(openapi3.Tags, 0, 10),
	}

	return &OpenAPI{doc: t}
}

func (doc *OpenAPI) Doc() *openapi3.T { return doc.doc }

// SaveAs 保存为 yaml 或 json 文件
//
// 根据后缀名名确定保存的文件类型，目前仅支持 json 和 yaml。
func (doc *OpenAPI) SaveAs(path string) error {
	var m func(any) ([]byte, error)
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		m = yaml.Marshal
	case ".json":
		m = func(v any) ([]byte, error) {
			return json.MarshalIndent(v, "", "\t")
		}
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
	cloneMap(d.Paths, doc.Doc().Paths)
	if d.Components != nil {
		cloneMap(d.Components.Schemas, doc.Doc().Components.Schemas)
		cloneMap(d.Components.Parameters, doc.Doc().Components.Parameters)
		cloneMap(d.Components.Headers, doc.Doc().Components.Headers)
		cloneMap(d.Components.RequestBodies, doc.Doc().Components.RequestBodies)
		cloneMap(d.Components.Responses, doc.Doc().Components.Responses)
		cloneMap(d.Components.SecuritySchemes, doc.Doc().Components.SecuritySchemes)
		cloneMap(d.Components.Examples, doc.Doc().Components.Examples)
		cloneMap(d.Components.Links, doc.Doc().Components.Links)
		cloneMap(d.Components.Callbacks, doc.Doc().Components.Callbacks)
	}
}

func cloneMap[K comparable, V any](src, dest map[K]V) {
	for k, v := range src {
		dest[k] = v
	}
}
