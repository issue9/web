// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

// Package codecs 提供编解码的实现
package codecs

import (
	"bytes"
	"encoding/gob"
	"encoding/json/v2"
	"encoding/xml"
	"io"

	"github.com/fxamacker/cbor/v2"
	"github.com/goccy/go-yaml"

	"github.com/issue9/web"
)

type (
	Codec = web.Codec

	codec struct {
		marshal   web.MarshalFunc
		unmarshal web.UnmarshalFunc
	}

	nopCodec  struct{}
	cborCodec struct{}
	gobCodec  struct{}
	jsonCodec struct{}
	xmlCodec  struct{}
	yamlCodec struct{}
)

// NewCodec 声明一个媒体类型
func NewCodec(m web.MarshalFunc, u web.UnmarshalFunc) Codec {
	return &codec{
		marshal:   m,
		unmarshal: u,
	}
}
func (m *codec) Marshal(ctx *web.Context, v any) ([]byte, error) { return m.marshal(ctx, v) }
func (m *codec) Unmarshal(r io.Reader, v any) error              { return m.unmarshal(r, v) }

// NewNOP 创建一个空的 [MediaType] 实例
//
// 调用创建对象的方法都将返回 [negotiation.ErrUnsupported] 错误，名称也都为空。
func NewNOP() Codec                                            { return nopCodec{} }
func (nopCodec) Marshal(_ *web.Context, v any) ([]byte, error) { return nil, ErrUnsupported() }
func (nopCodec) Unmarshal(r io.Reader, v any) error            { return ErrUnsupported() }

// NewCBOR 提供对 [CBOR] 编码的支持
//
// [CBOR]: https://www.rfc-editor.org/rfc/rfc8949.html
func NewCBOR() Codec                                            { return cborCodec{} }
func (cborCodec) Marshal(_ *web.Context, v any) ([]byte, error) { return cbor.Marshal(v) }
func (cborCodec) Unmarshal(r io.Reader, v any) error            { return cbor.NewDecoder(r).Decode(v) }

// NewGOB 提供对 [GOB] 编码的支持
//
// [GOB]: https://pkg.go.dev/encoding/gob
func NewGOB() Codec                                 { return gobCodec{} }
func (gobCodec) Unmarshal(r io.Reader, v any) error { return gob.NewDecoder(r).Decode(v) }
func (gobCodec) Marshal(_ *web.Context, v any) ([]byte, error) {
	w := new(bytes.Buffer)
	if err := gob.NewEncoder(w).Encode(v); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// NewJSON 提供对 JSON 编码的支持
func NewJSON() Codec                                            { return jsonCodec{} }
func (jsonCodec) Marshal(_ *web.Context, v any) ([]byte, error) { return json.Marshal(v) }
func (jsonCodec) Unmarshal(r io.Reader, v any) error            { return json.UnmarshalRead(r, v) }

// NewXML 提供对 XML 编码的支持
func NewXML() Codec                                            { return xmlCodec{} }
func (xmlCodec) Marshal(_ *web.Context, v any) ([]byte, error) { return xml.Marshal(v) }
func (xmlCodec) Unmarshal(r io.Reader, v any) error            { return xml.NewDecoder(r).Decode(v) }

// NewYAML 提供对 YAML 编码的支持
func NewYAML() Codec                                            { return yamlCodec{} }
func (yamlCodec) Marshal(_ *web.Context, v any) ([]byte, error) { return yaml.Marshal(v) }
func (yamlCodec) Unmarshal(r io.Reader, v any) error            { return yaml.NewDecoder(r).Decode(v) }
