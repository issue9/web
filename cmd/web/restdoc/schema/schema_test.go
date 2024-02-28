// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package schema

import (
	"go/types"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v4"
)

func TestBuildBasicType(t *testing.T) {
	a := assert.New(t, false)

	ref, ok := buildBasicType(types.Typ[types.Int])
	a.True(ok).Equal(ref.Value.Type, openapi3.TypeInteger)
}
