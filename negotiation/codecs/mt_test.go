// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package codecs

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v6"

	"github.com/issue9/web/negotiation"
)

func TestNewPatcherMediaType(t *testing.T) {
	a := assert.New(t, false)

	codec := NewJSON()
	mt := NewPatcherMediaType(codec, "name")
	a.Equal(mt.Codec, codec).
		StrictEqual(mt.Method(), http.MethodPatch).
		StrictEqual(mt.Name(), "name").
		Empty(mt.ProblemName()).
		StrictEqual(mt.Method(), http.MethodPatch).
		StrictEqual(mt.PatchKind(), negotiation.PatchKindPatcher)
}
