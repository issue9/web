// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package negotiation_test

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v10/header"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/qheader"
	"github.com/issue9/web/negotiation"
	"github.com/issue9/web/negotiation/codecs"
)

func TestNewMediaType(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		negotiation.NewMediaType[*web.Context](nil, "", "", "", negotiation.PatchKindNone)
	}, "参数 codec 不能为 nil")

	a.PanicString(func() {
		negotiation.NewMediaType[*web.Context](codecs.NewJSON(), "", "", "", negotiation.PatchKindNone)
	}, "参数 name 不能为空")

	codec := codecs.NewJSON()
	nego := negotiation.NewMediaType(codec, "name", "problem", http.MethodPatch, negotiation.PatchKindPatcher)
	a.Equal(nego.Codec, codec).
		StrictEqual(nego.Method(), http.MethodPatch).
		StrictEqual(nego.Name(), "name").
		StrictEqual(nego.ProblemName(), "problem").
		StrictEqual(nego.Method(), http.MethodPatch).
		StrictEqual(nego.PatchKind(), negotiation.PatchKindPatcher)
}

func TestNegotiation_Accept(t *testing.T) {
	a := assert.New(t, false)
	nego := negotiation.New[*web.Context](language.SimplifiedChinese, nil)
	a.NotNil(nego)

	item, err := nego.Accept(header.JSON)
	a.NotError(err).Nil(item)

	item, err = nego.Accept("")
	a.NotError(err).Nil(item)

	nego = negotiation.New[*web.Context](language.SimplifiedChinese, nil)
	a.NotNil(nego)
	nego.AddCodec(codecs.NewJSON(), header.JSON, "", "", negotiation.PatchKindNone).
		AddCodec(codecs.NewXML(), header.Plain, "text/plain+problem", "", negotiation.PatchKindNone).
		AddCodec(codecs.NewXML(), header.XML, "", http.MethodPatch, negotiation.PatchKindNone)

	t.Run("JSON", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.Accept(header.JSON)
		a.NotError(err).NotNil(item).
			NotNil(item.Marshal).
			StrictEqual(item.Name(), header.JSON).
			StrictEqual(item.ProblemName(), header.JSON)
	})

	t.Run("*/*", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.Accept("*/*")
		a.NotError(err).NotNil(item).
			NotNil(item.Marshal).
			StrictEqual(item.Name(), header.JSON)
	})

	t.Run("空参数，结果同 */*", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.Accept("")
		a.NotError(err).NotNil(item).
			NotNil(item.Marshal).
			StrictEqual(item.Name(), header.JSON)
	})

	t.Run("*/*,text/plain", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.Accept("*/*,text/plain")
		a.NotError(err).NotNil(item).
			NotNil(item.Marshal).
			StrictEqual(item.Name(), header.Plain).
			StrictEqual(item.ProblemName(), "text/plain+problem")
	})

	t.Run("json 被禁，但是 */* 匹配了 text/plain", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.Accept(header.JSON + ";q=0,*/*;q=1")
		a.NotError(err).StrictEqual(item.Name(), header.Plain)
	})

	t.Run("json 被禁", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.Accept(header.JSON + ";q=0")
		a.NotError(err).Nil(item)
	})

	t.Run("json 和 text/* 被禁，XML 可用", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.Accept("text/*;q=0,application/json;q=0,*/*;q=1")
		a.NotError(err).NotNil(item).StrictEqual(item.Name(), header.XML)
	})

	t.Run("font/wottf;q=x.9", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.Accept("font/wottf;q=x.9")
		a.Error(err).Nil(item)
	})

	t.Run("font/wottf", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.Accept("font/wottf")
		a.NotError(err).Nil(item)
	})
}

func TestNegotiation_ContentType(t *testing.T) {
	a := assert.New(t, false)

	nego := negotiation.New[*web.Context](language.SimplifiedChinese, nil)
	a.NotNil(nego)
	nego.AddMediaType(
		negotiation.NewMediaType[*web.Context](codecs.NewJSON(), header.XML, "", "", negotiation.PatchKindNone),
		negotiation.NewMediaType[*web.Context](codecs.NewJSON(), header.JSON, "", http.MethodPatch, negotiation.PatchKindNone),
		negotiation.NewMediaType[*web.Context](codecs.NewJSON(), header.CBOR, "", "QUERY", negotiation.PatchKindNone),
	)

	t.Run(";;;", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType(";;;", "")
		a.Error(err).Nil(f).Nil(e)
	})

	t.Run("不存在的 mediaType", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType("not-exists; charset=utf-8", "")
		a.Equal(err, localeutil.Error("not found serialization function for %s", "not-exists")).Nil(f).Nil(e)
	})

	t.Run("xml;charset=utf-8", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType(qheader.BuildContentType(header.XML, header.UTF8), "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("xml;charset=gb2312", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType(qheader.BuildContentType(header.XML, "gb2312"), "")
		a.NotError(err).NotNil(f).NotNil(e)
	})

	t.Run("xml;charset=not-exists", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType(qheader.BuildContentType(header.XML, "not-exists"), "")
		a.Error(err).Nil(f).Nil(e)
	})

	t.Run("xml;charset=UTF-8;params", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType("application/xml; charset=UTF-8;p1=k1;p2=k2", "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("xml;charset=", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType(qheader.BuildContentType(header.XML, ""), "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("没有 charset", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType("application/xml;", "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("无 ;charset", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType(header.XML, "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	// Patch 在没有匹配 PATCH 的请求的情况下，也会匹配普通项
	t.Run("xml;charset=utf-8;method=PATCH", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType(qheader.BuildContentType(header.XML, header.UTF8), http.MethodPatch)
		a.NotError(err).NotNil(f).Nil(e)
	})

	// Patch 匹配 PATCH 项
	t.Run("json;charset=utf-8;method=PATCH", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType(qheader.BuildContentType(header.JSON, header.UTF8), http.MethodPatch)
		a.NotError(err).NotNil(f).Nil(e)
	})

	// 普通请求无法匹配 Patch 项
	t.Run("json;charset=utf-8;method=GET", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType(qheader.BuildContentType(header.JSON, header.UTF8), http.MethodGet)
		a.Error(err).Nil(f).Nil(e)
	})

	// 普通请求匹配普通项
	t.Run("xml;charset=utf-8;method=GET", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.ContentType(qheader.BuildContentType(header.XML, header.UTF8), http.MethodGet)
		a.NotError(err).NotNil(f).Nil(e)
	})
}

func TestNegotiation_AddCodec(t *testing.T) {
	a := assert.New(t, false)
	nego := negotiation.New[*web.Context](language.SimplifiedChinese, nil)

	a.NotPanic(func() { nego.AddCodec(codecs.NewJSON(), header.JSON, "", "", negotiation.PatchKindNone) })

	a.PanicString(func() {
		nego.AddCodec(codecs.NewJSON(), header.JSON, "", "", negotiation.PatchKindNone)
	}, "存在重复的项 "+header.JSON)

	t.Run("AcceptHeader", func(t *testing.T) {
		mt := negotiation.NewMediaType[*web.Context](codecs.NewJSON(), header.Plain, "", "", negotiation.PatchKindNone)
		e := negotiation.New[*web.Context](language.SimplifiedChinese, nil).AddMediaType(mt)
		a.Equal(e.AcceptHeader(), header.Plain).
			Empty(e.AcceptPatchHeader()).
			Empty(e.AcceptQueryHeader())
	})

	t.Run("AcceptPatchHeader", func(t *testing.T) {
		mt := negotiation.NewMediaType[*web.Context](codecs.NewJSON(), header.Plain, "", http.MethodPatch, negotiation.PatchKindNone)
		e := negotiation.New[*web.Context](language.SimplifiedChinese, nil).AddMediaType(mt)
		a.Equal(e.AcceptPatchHeader(), header.Plain).
			Empty(e.AcceptHeader()).
			Empty(e.AcceptQueryHeader())
	})

	t.Run("AcceptQueryHeader", func(t *testing.T) {
		mt := negotiation.NewMediaType[*web.Context](codecs.NewJSON(), header.Plain, "", "QUERY", negotiation.PatchKindNone)
		e := negotiation.New[*web.Context](language.SimplifiedChinese, nil).AddMediaType(mt)
		a.Equal(e.AcceptQueryHeader(), header.Plain).
			Empty(e.AcceptHeader()).
			Empty(e.AcceptPatchHeader())
	})

	t.Run("Accept", func(t *testing.T) {
		mt1 := negotiation.NewMediaType[*web.Context](codecs.NewJSON(), "application/sql", "", "QUERY", negotiation.PatchKindNone)
		mt2 := negotiation.NewMediaType[*web.Context](codecs.NewJSON(), "text/plain", "", "", negotiation.PatchKindNone)
		mt3 := negotiation.NewMediaType[*web.Context](codecs.NewJSON(), "application/json-patch+json", "", http.MethodPatch, negotiation.PatchKindNone)
		mt4 := negotiation.NewMediaType[*web.Context](codecs.NewJSON(), "application/json-merge+json", "", http.MethodPatch, negotiation.PatchKindNone)
		mt5 := negotiation.NewMediaType[*web.Context](codecs.NewJSON(), "application/cbor", "", "", negotiation.PatchKindNone)
		e := negotiation.New[*web.Context](language.SimplifiedChinese, nil).AddMediaType(mt1, mt2, mt3, mt4, mt5)
		a.Equal(e.AcceptQueryHeader(), "application/sql").
			Equal(e.AcceptHeader(), "text/plain, application/cbor").
			Equal(e.AcceptPatchHeader(), "application/json-patch+json, application/json-merge+json")
	})
}
