// SPDX-License-Identifier: MIT

package schema

import (
	"go/types"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v3"
)

func TestBuildBasicType(t *testing.T) {
	a := assert.New(t, false)

	ref, ok := buildBasicType(types.Typ[types.Int])
	a.True(ok).Equal(ref.Value.Type, openapi3.TypeInteger)
}
