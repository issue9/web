// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package codec

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
	normalName  string
	problemName string
	marshal     web.MarshalFunc
	unmarshal   web.UnmarshalFunc
}

// NewMediaType 声明一个媒体类型
//
// name 媒体类型的名称，比如 application/json；
// problem 媒体类型非正常状态下的子类型，比如 application/problem+json；
func NewMediaType(name, problem string, m web.MarshalFunc, u web.UnmarshalFunc) MediaType {
	if problem == "" {
		problem = name
	}

	return &mediaType{
		normalName:  name,
		problemName: problem,
		marshal:     m,
		unmarshal:   u,
	}
}

func (m *mediaType) Name() string { return m.normalName }

func (m *mediaType) ProblemName() string { return m.problemName }

func (m *mediaType) Marshal(ctx *web.Context, v any) ([]byte, error) { return m.marshal(ctx, v) }

func (m *mediaType) Unmarshal(r io.Reader, v any) error { return m.unmarshal(r, v) }

// ForkMediaType 使用已有的 [MediaType] 的编解码方法构建新的 [MediaType]
func ForkMediaType(mt MediaType, normal, problem string) MediaType {
	return NewMediaType(normal, problem, mt.Marshal, mt.Unmarshal)
}

// NewNOP 创建一个空的 [MeidaType] 实例
//
// 调用创建对象的方法都将返回 [codec.ErrUnsupported] 错误，名称也都为空。
func NewNOP() MediaType {
	return NewMediaType("nop", "",
		func(_ *web.Context, _ any) ([]byte, error) { return nil, ErrUnsupported() },
		func(_ io.Reader, _ any) error { return ErrUnsupported() },
	)
}

// NewCBOR 提供对 [CBOR] 编码的支持
//
// [CBOR]: https://www.rfc-editor.org/rfc/rfc8949.html
func NewCBOR() MediaType {
	return NewMediaType("application/cbor", "application/problem+cbor",
		func(_ *web.Context, v any) ([]byte, error) { return cbor.Marshal(v) },
		func(r io.Reader, v any) error { return cbor.NewDecoder(r).Decode(v) },
	)
}

// NewGOB 提供对 [GOB] 编码的支持
//
// [GOB]: https://pkg.go.dev/encoding/gob
func NewGOB() MediaType {
	return NewMediaType(header.OctetStream, "",
		func(ctx *web.Context, a any) ([]byte, error) {
			w := new(bytes.Buffer)
			enc := gob.NewEncoder(w)
			if err := enc.Encode(a); err != nil {
				return nil, err
			}

			return w.Bytes(), nil
		},
		func(r io.Reader, a any) error { return gob.NewDecoder(r).Decode(a) },
	)
}

// NewJSON 提供对 JSON 编码的支持
func NewJSON() MediaType {
	return NewMediaType(header.JSON, "application/problem+json",
		func(ctx *web.Context, a any) ([]byte, error) { return json.Marshal(a) },
		func(r io.Reader, a any) error { return json.UnmarshalRead(r, a) },
	)
}

// NewXML 提供对 XML 编码的支持
func NewXML() MediaType {
	return NewMediaType(header.XML, "application/problem+xml",
		func(ctx *web.Context, a any) ([]byte, error) { return xml.Marshal(a) },
		func(r io.Reader, a any) error { return xml.NewDecoder(r).Decode(a) },
	)
}

// NewYAML 提供对 YAML 编码的支持
func NewYAML() MediaType {
	return NewMediaType(header.YAML, "application/problem+yaml",
		func(ctx *web.Context, a any) ([]byte, error) { return yaml.Marshal(a) },
		func(r io.Reader, a any) error { return yaml.NewDecoder(r).Decode(a) },
	)
}
