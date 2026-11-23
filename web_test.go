// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json/v2"
	"encoding/xml"
	"errors"
	"io"
	"slices"

	"github.com/issue9/assert/v6"
	"github.com/issue9/config"
	"github.com/issue9/mux/v10/header"
	"golang.org/x/text/language"

	"github.com/issue9/web/negotiation"
)

type (
	gzipImpl    struct{}
	deflateImpl struct{}

	codecImpl struct {
		m negotiation.MarshalFunc[*Context]
		u negotiation.UnmarshalFunc
	}
)

func newGzip() negotiation.Compressor                           { return &gzipImpl{} }
func (gzipImpl) Name() string                                   { return "gzip" }
func (gzipImpl) NewDecoder(r io.Reader) (io.ReadCloser, error)  { return gzip.NewReader(r) }
func (gzipImpl) NewEncoder(w io.Writer) (io.WriteCloser, error) { return gzip.NewWriter(w), nil }

func newDeflate() negotiation.Compressor                          { return &deflateImpl{} }
func (deflateImpl) Name() string                                  { return "deflate" }
func (deflateImpl) NewDecoder(r io.Reader) (io.ReadCloser, error) { return flate.NewReader(r), nil }
func (deflateImpl) NewEncoder(w io.Writer) (io.WriteCloser, error) {
	return flate.NewWriter(w, flate.DefaultCompression)
}

func newCodec(m negotiation.MarshalFunc[*Context], u negotiation.UnmarshalFunc) Codec {
	return &codecImpl{m: m, u: u}
}
func (m *codecImpl) Marshal(ctx *Context, v any) ([]byte, error) { return m.m(ctx, v) }
func (m *codecImpl) Unmarshal(r io.Reader, v any) error          { return m.u(r, v) }

func marshalTest(_ *Context, v any) ([]byte, error) {
	switch vv := v.(type) {
	case error:
		return nil, vv
	default:
		return nil, errors.ErrUnsupported // 测试环境，直接返回 errors.ErrUnsupported 了
	}
}

func unmarshalTest(io.Reader, any) error            { return errors.ErrUnsupported }
func marshalJSON(_ *Context, v any) ([]byte, error) { return json.Marshal(v) }
func unmarshalJSON(r io.Reader, v any) error        { return json.UnmarshalRead(r, v) }
func marshalXML(_ *Context, v any) ([]byte, error)  { return xml.Marshal(v) }
func unmarshalXML(r io.Reader, v any) error         { return xml.NewDecoder(r).Decode(v) }

func newNegotiation(a *assert.Assertion, conf *config.Config) *Negotiation {
	nego := negotiation.New[*Context](language.SimplifiedChinese, conf)
	a.NotNil(nego)

	nego.AddCompressor(newGzip()).
		AddCompressor(newDeflate()).
		//AddCompressor(nil).
		AddCodec(newCodec(marshalJSON, unmarshalJSON), header.JSON, "application/problem+json", "", negotiation.PatchKindNone).
		AddCodec(newCodec(marshalXML, unmarshalXML), header.XML, "application/problem+xml", "", negotiation.PatchKindNone).
		AddCodec(newCodec(marshalTest, unmarshalTest), "application/test", "application/problem+test", "", negotiation.PatchKindNone)

	a.Length(slices.Collect(nego.MediaTypes()), 3).
		StrictEqual(conf, nego.Config()).
		StrictEqual(nego.Language(), language.SimplifiedChinese)

	return nego
}
