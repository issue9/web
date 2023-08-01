// SPDX-License-Identifier: MIT

package schema

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v3"
)

func TestWrap(t *testing.T) {
	a := assert.New(t, false)

	ref := openapi3.NewSchemaRef("ref", openapi3.NewBoolSchema())
	ref2 := wrap(ref, "", "", nil, false)
	a.Equal(ref2, ref)

	ref2 = wrap(ref, "", "123", nil, false)
	a.NotEqual(ref2, ref).
		Equal(ref2.Value.AllOf[0].Value, ref.Value).
		Equal(ref2.Value.Description, "123")

	ref2 = wrap(ref, "", "123", nil, true)
	a.NotEqual(ref2, ref).
		Equal(ref2.Value.AllOf[0].Value, ref.Value).
		Equal(ref2.Value.Description, "123").
		True(ref2.Value.Nullable)
}
