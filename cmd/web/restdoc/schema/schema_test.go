// SPDX-License-Identifier: MIT

package schema

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v3"
)

func TestBuildBasicType(t *testing.T) {
	a := assert.New(t, false)

	ref, ok := buildBasicType("int")
	a.True(ok).Equal(ref.Value.Type, openapi3.TypeInteger)

	ref, ok = buildBasicType("map")
	a.True(ok).Equal(ref.Value.Type, openapi3.TypeObject)

	ref, ok = buildBasicType("{}")
	a.True(ok).Nil(ref)
}
