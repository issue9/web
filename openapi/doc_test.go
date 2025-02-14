// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web"
)

func TestParameterizedDescription(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	p := s.Locale().NewPrinter(language.SimplifiedChinese)

	t.Run("panic", func(*testing.T) {
		d := New(s, web.Phrase("title"))

		d.addOperation("GET", "/users/1", "", &Operation{
			d:         d,
			Responses: map[string]*Response{"200": {Body: &Schema{Type: TypeNumber}}},
			ID:        "id1",
		})
		d.AppendDescriptionParameter("id1", web.Phrase("item1"), web.Phrase("item2"))
		a.PanicString(func() {
			d.build(p, language.SimplifiedChinese, nil)
		}, "接口 id1 未指定 Description 内容").
			Length(d.parameterizedDesc, 1)
	})

	t.Run("not parameterizedDescription", func(*testing.T) {
		d := New(s, web.Phrase("title"))

		d.addOperation("GET", "/users/2", "", &Operation{
			d:           d,
			Responses:   map[string]*Response{"200": {Body: &Schema{Type: TypeNumber}}},
			ID:          "id2",
			Description: web.Phrase("desc"),
		})

		d.AppendDescriptionParameter("id2", web.Phrase("item1"), web.Phrase("item2"))
		a.PanicString(func() {
			d.build(p, language.SimplifiedChinese, nil)
		}, "向 [id2] 注册的参数并未使用")
	})

	t.Run("parameterizedDescription", func(*testing.T) {
		d := New(s, web.Phrase("title"))

		d.addOperation("GET", "/users/3", "", &Operation{
			d:           d,
			Responses:   map[string]*Response{"200": {Body: &Schema{Type: TypeNumber}}},
			ID:          "id3",
			Description: ParameterizedDoc("desc %s", nil),
		})

		d.AppendDescriptionParameter("id3", web.Phrase("item1"), web.Phrase("item2"))
		r := d.build(p, language.SimplifiedChinese, nil)
		a.Equal(r.Paths.GetPair("/users/3").Value.obj.Get.Description, "desc item1\nitem2\n")
	})

	t.Run("invalid format", func(*testing.T) {
		a.PanicString(func() {
			ParameterizedDoc("desc", nil)
		}, "参数 format 必须包含 %s")
	})
}
