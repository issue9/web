// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

// Version 支持的 openapi 版本
const Version = "3.1.0"

const (
	InPath   = "path"
	InQuery  = "query"
	InHeader = "header"
	InCookie = "cookie"

	TypeString  = "string"
	TypeObject  = "object"
	TypeNull    = "null"
	TypeBoolean = "boolean"
	TypeArray   = "array"
	TypeNumber  = "number"
	TypeInteger = "integer"

	FormatInt32    = "int32"
	FormatInt64    = "int64"
	FormatFloat    = "float"
	FormatDouble   = "double"
	FormatPassword = "password"

	SecuritySchemeTypeHTTP          = "http"
	SecuritySchemeTypeAPIKey        = "apiKey"
	SecuritySchemeTypeMutualTLS     = "mutualTLS"
	SecuritySchemeTypeOAuth2        = "oauth2"
	SecuritySchemeTypeOpenIDConnect = "openIdConnect"

	// CommentTag 可提取翻译内容的结构体标签名称
	CommentTag = "comment"
)
