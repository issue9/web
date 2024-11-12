// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"encoding/json"
	"testing"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web"
)

var (
	_ json.Marshaler = &renderer[int]{}
	_ yaml.Marshaler = &renderer[int]{}
)

func TestRef_build(t *testing.T) {
	a := assert.New(t, false)
	p := newPrinter(a, language.SimplifiedChinese)

	ref := &Ref{}
	a.PanicString(func() {
		ref.build(p, "schemas")
	}, "ref 不能为空")

	ref = &Ref{
		Ref:         "ref",
		Summary:     web.Phrase("lang"),
		Description: web.Phrase("desc"),
	}
	a.Equal(ref.build(p, "schemas"), &refRenderer{Ref: "#/components/schemas/ref", Summary: "简体", Description: "desc"})
}

type object struct {
	ID int
}

func TestRenderer(t *testing.T) {
	a := assert.New(t, false)
	p := newPrinter(a, language.SimplifiedChinese)

	a.PanicString(func() {
		newRenderer[int](nil, nil)
	}, "ref 和 obj 不能同时为 nil")

	ref := &Ref{Ref: "ref"}

	r := newRenderer[object](ref.build(p, "schemas"), &object{ID: 2})
	a.Equal(r.ref.Ref, "#/components/schemas/ref").
		Empty(r.ref.Summary).
		NotNil(r.obj)

	// JSON
	bs, err := json.Marshal(r)
	a.NotError(err).Equal(string(bs), `{"$ref":"#/components/schemas/ref"}`)

	// YAML
	bs, err = yaml.Marshal(r)
	a.NotError(err).Equal(string(bs), "$ref: '#/components/schemas/ref'\n")

	// ref = nil

	r = newRenderer(nil, &object{ID: 2})
	a.Nil(r.ref).NotNil(r.obj)

	// JSON
	bs, err = json.Marshal(r)
	a.NotError(err).Equal(string(bs), `{"ID":2}`)

	// YAML
	bs, err = yaml.Marshal(r)
	a.NotError(err).Equal(string(bs), "id: 2\n")
}
