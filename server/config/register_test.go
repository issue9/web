// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert/v6"

	"github.com/issue9/web/negotiation"
	"github.com/issue9/web/server"
)

func TestRegisterFileSerializer(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		RegisterFileSerializer("new", server.JSONMarshal, server.JSONUnmarshal, ".json")
	}, "扩展名 .json 已经注册到 json")

	RegisterFileSerializer("new", server.JSONMarshal, server.JSONUnmarshal, ".js")
	v, f := fileSerializerFactory.get("new")
	a.True(f).NotNil(v).Equal(v.exts, []string{".js"})

	a.TB().Cleanup(func() {
		delete(fileSerializerFactory.items, "new")
	})
}

func TestRegisterCodec(t *testing.T) {
	a := assert.New(t, false)

	v, f := codecFactory.get("json")
	a.True(f).NotNil(v)

	RegisterCodec(negotiation.NewJSON(), "json2")
	v, f = codecFactory.get("json2")
	a.True(f).NotNil(v)

	a.PanicString(func() {
		RegisterCodec(negotiation.NewJSON(), "json2")
	}, "存在同名的媒体类型:json2")
}
