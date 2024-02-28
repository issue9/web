// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package schema

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v4"
)

func TestNewPathSchema(t *testing.T) {
	a := assert.New(t, false)
	s, err := NewPath("int")
	a.NotError(err).NotNil(s).Empty(s.Ref).Equal(s.Value.Type, openapi3.TypeInteger)

	s, err = NewPath("boolean")
	a.NotError(err).NotNil(s).Empty(s.Ref).Equal(s.Value.Type, openapi3.TypeBoolean)

	s, err = NewPath("str")
	a.NotError(err).NotNil(s).Empty(s.Ref).Equal(s.Value.Type, openapi3.TypeString)

	s, err = NewPath("float32")
	a.NotError(err).NotNil(s).Empty(s.Ref).Equal(s.Value.Type, openapi3.TypeNumber)

	s, err = NewPath("id")
	a.NotError(err).NotNil(s).Empty(s.Ref).
		Equal(s.Value.Type, openapi3.TypeInteger).
		Equal(*s.Value.Min, 1)

	s, err = NewPath("\\s+")
	a.NotError(err).NotNil(s).Empty(s.Ref).
		Empty(s.Value.Type).
		Equal(s.Value.Pattern, "\\s+")

	s, err = NewPath("(\\)+") // 无效的正则表达式
	a.Error(err).Nil(s)
}
