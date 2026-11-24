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

// NewPatcherMediaType 声明 patchKind 值为 [negotiation.PatchKindPatcher] 的 [MediaType]
func NewPatcherMediaType(codec Codec, name string) *MediaType {
	return NewMediaType(codec, name, "", http.MethodPatch, negotiation.PatchKindPatcher)
}

// NewPatcherMediaType 声明 patchKind 值为 [negotiation.PatchKindMerger] 的 [MediaType]
func NewMergerMediaType(codec Codec, name string) *MediaType {
	return NewMediaType(codec, name, "", http.MethodPatch, negotiation.PatchKindMerger)
}
