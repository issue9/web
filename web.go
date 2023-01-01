// SPDX-License-Identifier: MIT

// Package web 模块化的 web 框架
//
// NOTE: 所有以 Internal 开头的公开函数，表示这个函数是仅模块可见的。
package web

import (
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"github.com/issue9/query/v3"
	"golang.org/x/text/message"

	"github.com/issue9/web/internal/base"
	"github.com/issue9/web/server"
)

// Version 当前框架的版本
const Version = "0.61.0"

// 预定义的 Problem id 值
const (
	// 400
	ProblemBadRequest                   = "400"
	ProblemUnauthorized                 = "401"
	ProblemPaymentRequired              = "402"
	ProblemForbidden                    = "403"
	ProblemNotFound                     = "404"
	ProblemMethodNotAllowed             = "405"
	ProblemNotAcceptable                = "406"
	ProblemProxyAuthRequired            = "407"
	ProblemRequestTimeout               = "408"
	ProblemConflict                     = "409"
	ProblemGone                         = "410"
	ProblemLengthRequired               = "411"
	ProblemPreconditionFailed           = "412"
	ProblemRequestEntityTooLarge        = "413"
	ProblemRequestURITooLong            = "414"
	ProblemUnsupportedMediaType         = "415"
	ProblemRequestedRangeNotSatisfiable = "416"
	ProblemExpectationFailed            = "417"
	ProblemTeapot                       = "418"
	ProblemMisdirectedRequest           = "421"
	ProblemUnprocessableEntity          = "422"
	ProblemLocked                       = "423"
	ProblemFailedDependency             = "424"
	ProblemTooEarly                     = "425"
	ProblemUpgradeRequired              = "426"
	ProblemPreconditionRequired         = "428"
	ProblemTooManyRequests              = "429"
	ProblemRequestHeaderFieldsTooLarge  = "431"
	ProblemUnavailableForLegalReasons   = "451"

	// 500
	ProblemInternalServerError           = "500"
	ProblemNotImplemented                = "501"
	ProblemBadGateway                    = "502"
	ProblemServiceUnavailable            = "503"
	ProblemGatewayTimeout                = "504"
	ProblemHTTPVersionNotSupported       = "505"
	ProblemVariantAlsoNegotiates         = "506"
	ProblemInsufficientStorage           = "507"
	ProblemLoopDetected                  = "508"
	ProblemNotExtended                   = "510"
	ProblemNetworkAuthenticationRequired = "511"
)

type (
	Server         = server.Server
	Context        = server.Context
	Options        = server.Options
	MiddlewareFunc = server.MiddlewareFunc
	Middleware     = server.Middleware
	HandlerFunc    = server.HandlerFunc
	Router         = server.Router
	Responser      = server.Responser
	ResponserFunc  = server.ResponserFunc
	CTXSanitizer   = server.CTXSanitizer
	Rule           = server.Rule
	Validation     = server.Validation
	Validator      = server.Validator
	ValidateFunc   = server.ValidateFunc
	Logger         = logs.Logger

	// QueryUnmarshaler 对查询参数的解析接口
	QueryUnmarshaler = query.Unmarshaler

	// LocaleStringer 本地化字符串需要实在的接口
	//
	// 部分 error 返回可能也实现了该接口。
	LocaleStringer = localeutil.LocaleStringer
)

func NewServer(name, version string, o *Options) (*Server, error) {
	return server.New(name, version, o)
}

// Phrase 生成本地化的语言片段
func Phrase(key message.Reference, v ...any) LocaleStringer { return localeutil.Phrase(key, v...) }

// NewRule 新建验证规则
func NewRule(msg LocaleStringer, v Validator) *Rule { return server.NewRule(msg, v) }

// NewRuleFunc 新建验证规则
func NewRuleFunc(msg LocaleStringer, f func(any) bool) *Rule {
	return server.NewRuleFunc(msg, f)
}

// StackError 为 err 带上调用信息
//
// 位置从调用 StackError 开始。
// 如果 err 为 nil，则返回 nil，如果 err 本身就为 StackError 返回的类型，则原样返回。
//
// 如果需要输出调用堆栈信息，需要指定 %+v 标记。
func StackError(err error) error { return base.StackError(err) }
