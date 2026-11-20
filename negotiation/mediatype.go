// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package negotiation

import (
	"bytes"
	"encoding/gob"
	"encoding/json/v2"
	"encoding/xml"
	"io"

	"github.com/fxamacker/cbor/v2"
	"github.com/goccy/go-yaml"
	"github.com/issue9/mux/v10/header"

	"github.com/issue9/web"
)

type MediaType = web.MediaType

type mediaType struct {
	name        string
	problemName string
	method      string
	marshal     web.MarshalFunc
	unmarshal   web.UnmarshalFunc
}

// NewMediaType 声明一个媒体类型
//
// name 媒体类型的名称，比如 application/json；
// problem 媒体类型非正常状态下的子类型，比如 application/problem+json；
// method 绑定的请求方法，具体可用的值可参考 [web.MediaType.Method]；
func NewMediaType(name, problem, method string, m web.MarshalFunc, u web.UnmarshalFunc) MediaType {
	if problem == "" {
		problem = name
	}

	return &mediaType{
		name:        name,
		problemName: problem,
		method:      method,
		marshal:     m,
		unmarshal:   u,
	}
}

func (m *mediaType) Name() string                                    { return m.name }
func (m *mediaType) ProblemName() string                             { return m.problemName }
func (m *mediaType) Method() string                                  { return m.method }
func (m *mediaType) Marshal(ctx *web.Context, v any) ([]byte, error) { return m.marshal(ctx, v) }
func (m *mediaType) Unmarshal(r io.Reader, v any) error              { return m.unmarshal(r, v) }

type forkedMediaType struct {
	MediaType
	name        string
	problemName string
	method      string
}

// ForkMediaType 使用已有的 [MediaType] 的编解码方法构建新的 [MediaType]
func ForkMediaType(mt MediaType, name, problem, method string) MediaType {
	if problem == "" {
		problem = name
	}

	return &forkedMediaType{
		MediaType:   mt,
		name:        name,
		problemName: problem,
		method:      method,
	}
}

func (m *forkedMediaType) Name() string        { return m.name }
func (m *forkedMediaType) ProblemName() string { return m.problemName }
func (m *forkedMediaType) Method() string      { return m.method }

// NewNOP 创建一个空的 [MediaType] 实例
//
// 调用创建对象的方法都将返回 [negotiation.ErrUnsupported] 错误，名称也都为空。
func NewNOP(name, problem string) MediaType {
	return NewMediaType(name, problem, "",
		func(_ *web.Context, _ any) ([]byte, error) { return nil, ErrUnsupported() },
		func(_ io.Reader, _ any) error { return ErrUnsupported() },
	)
}

type cborMT struct{}

// NewCBOR 提供对 [CBOR] 编码的支持
//
// [CBOR]: https://www.rfc-editor.org/rfc/rfc8949.html
func NewCBOR() MediaType                                     { return &cborMT{} }
func (cborMT) Name() string                                  { return header.CBOR }
func (cborMT) ProblemName() string                           { return "application/problem+cbor" }
func (cborMT) Method() string                                { return "" }
func (cborMT) Marshal(_ *web.Context, v any) ([]byte, error) { return cbor.Marshal(v) }
func (cborMT) Unmarshal(r io.Reader, v any) error            { return cbor.NewDecoder(r).Decode(v) }

type gobMT struct{}

// NewGOB 提供对 [GOB] 编码的支持
//
// [GOB]: https://pkg.go.dev/encoding/gob
func NewGOB() MediaType                          { return &gobMT{} }
func (gobMT) Name() string                       { return header.OctetStream }
func (gobMT) ProblemName() string                { return "" }
func (gobMT) Method() string                     { return "" }
func (gobMT) Unmarshal(r io.Reader, v any) error { return gob.NewDecoder(r).Decode(v) }
func (gobMT) Marshal(_ *web.Context, v any) ([]byte, error) {
	w := new(bytes.Buffer)
	if err := gob.NewEncoder(w).Encode(v); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

type jsonMT struct{}

// NewJSON 提供对 JSON 编码的支持
func NewJSON() MediaType                                     { return &jsonMT{} }
func (jsonMT) Name() string                                  { return header.JSON }
func (jsonMT) ProblemName() string                           { return "application/problem+json" }
func (jsonMT) Method() string                                { return "" }
func (jsonMT) Marshal(_ *web.Context, v any) ([]byte, error) { return json.Marshal(v) }
func (jsonMT) Unmarshal(r io.Reader, v any) error            { return json.UnmarshalRead(r, v) }

type xmlMT struct{}

// NewXML 提供对 XML 编码的支持
func NewXML() MediaType                                     { return &xmlMT{} }
func (xmlMT) Name() string                                  { return header.XML }
func (xmlMT) ProblemName() string                           { return "application/problem+xml" }
func (xmlMT) Method() string                                { return "" }
func (xmlMT) Marshal(_ *web.Context, v any) ([]byte, error) { return xml.Marshal(v) }
func (xmlMT) Unmarshal(r io.Reader, v any) error            { return xml.NewDecoder(r).Decode(v) }

type yamlMT struct{}

// NewYAML 提供对 YAML 编码的支持
func NewYAML() MediaType                                     { return &yamlMT{} }
func (yamlMT) Name() string                                  { return header.YAML }
func (yamlMT) ProblemName() string                           { return "application/problem+yaml" }
func (yamlMT) Method() string                                { return "" }
func (yamlMT) Marshal(_ *web.Context, v any) ([]byte, error) { return yaml.Marshal(v) }
func (yamlMT) Unmarshal(r io.Reader, v any) error            { return yaml.NewDecoder(r).Decode(v) }
