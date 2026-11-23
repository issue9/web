// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package codecs

import (
	"net/http"

	"github.com/issue9/web"
	"github.com/issue9/web/negotiation"
)

type MediaType = negotiation.MediaType[*web.Context]

func NewMediaType(codec Codec, name, problem, method string, patchKind negotiation.PatchKind) *MediaType {
	return negotiation.NewMediaType(codec, name, problem, method, patchKind)
}

func NewPatcherMediaType(codec Codec, name, problem string) *MediaType {
	return NewMediaType(codec, name, problem, http.MethodPatch, negotiation.PatchKindPatcher)
}

func NewMergerMediaType(codec Codec, name, problem string) *MediaType {
	return NewMediaType(codec, name, problem, http.MethodPatch, negotiation.PatchKindMerger)
}
