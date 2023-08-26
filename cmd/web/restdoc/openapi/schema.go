// SPDX-License-Identifier: MIT

package openapi

import "github.com/getkin/kin-openapi/openapi3"

// AddSchema 添加一个 Schema 至 Components 中
func (doc *OpenAPI) AddSchema(ref string, schema *openapi3.Schema) {
	doc.schemaLocker.Lock()
	defer doc.schemaLocker.Unlock()

	doc.doc.Components.Schemas[ref] = openapi3.NewSchemaRef("", schema)
}

// GetSchema 从 Components 中查找 ref 引用的 Schema 定义
func (doc *OpenAPI) GetSchema(ref string) (*openapi3.SchemaRef, bool) {
	doc.schemaLocker.RLock()
	defer doc.schemaLocker.RUnlock()

	r, found := doc.doc.Components.Schemas[ref]
	return r, found
}
