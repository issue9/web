// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package codecs

import (
	"net/http"

	"github.com/issue9/web"
	"github.com/issue9/web/negotiation"
)

type MediaType = web.MediaType

// NewMediaType 创建新的媒体类型
//
// problem 错误状态下的媒体类型，如果为空则采用 name 值，其它参数可参考 [negotiation.NewMediaType]。
func NewMediaType(codec Codec, name, problem, method string, patchKind negotiation.PatchKind) *MediaType {
	if problem == "" {
		problem = name
	}
	return negotiation.NewMediaType(codec, name, problem, method, patchKind)
}

// NewPatcherMediaType 声明 patchKind 值为 [negotiation.PatchKindPatcher] 的 [MediaType]
func NewPatcherMediaType(codec Codec, name string) *MediaType {
	// 不要调用当前包的 NewMediaType，它会将 problem 改为与 name 相同的值。
	return negotiation.NewMediaType(codec, name, "", http.MethodPatch, negotiation.PatchKindPatcher)
}

// NewPatcherMediaType 声明 patchKind 值为 [negotiation.PatchKindMerger] 的 [MediaType]
func NewMergerMediaType(codec Codec, name string) *MediaType {
	return negotiation.NewMediaType(codec, name, "", http.MethodPatch, negotiation.PatchKindMerger)
}
