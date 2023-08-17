// SPDX-License-Identifier: MIT

// Package web 通用的 web 开发框架
package web

import (
	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/query/v3"

	"github.com/issue9/web/cache"
	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/internal/problems"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/server"
)

// Version 当前框架的版本
const Version = "0.80.0"

// 预定义的 Problem ID 值
const (
	// 特殊的值，当不想向用户展示 type 值时，该对象的 type 会被设置为该值。
	ProblemAboutBlank = problems.ProblemAboutBlank

	// 400
	ProblemBadRequest                   = problems.ProblemBadRequest
	ProblemUnauthorized                 = problems.ProblemUnauthorized
	ProblemPaymentRequired              = problems.ProblemPaymentRequired
	ProblemForbidden                    = problems.ProblemForbidden
	ProblemNotFound                     = problems.ProblemNotFound
	ProblemMethodNotAllowed             = problems.ProblemMethodNotAllowed
	ProblemNotAcceptable                = problems.ProblemNotAcceptable
	ProblemProxyAuthRequired            = problems.ProblemProxyAuthRequired
	ProblemRequestTimeout               = problems.ProblemRequestTimeout
	ProblemConflict                     = problems.ProblemConflict
	ProblemGone                         = problems.ProblemGone
	ProblemLengthRequired               = problems.ProblemLengthRequired
	ProblemPreconditionFailed           = problems.ProblemPreconditionFailed
	ProblemRequestEntityTooLarge        = problems.ProblemRequestEntityTooLarge
	ProblemRequestURITooLong            = problems.ProblemRequestURITooLong
	ProblemUnsupportedMediaType         = problems.ProblemUnsupportedMediaType
	ProblemRequestedRangeNotSatisfiable = problems.ProblemRequestedRangeNotSatisfiable
	ProblemExpectationFailed            = problems.ProblemExpectationFailed
	ProblemTeapot                       = problems.ProblemTeapot
	ProblemMisdirectedRequest           = problems.ProblemMisdirectedRequest
	ProblemUnprocessableEntity          = problems.ProblemUnprocessableEntity
	ProblemLocked                       = problems.ProblemLocked
	ProblemFailedDependency             = problems.ProblemFailedDependency
	ProblemTooEarly                     = problems.ProblemTooEarly
	ProblemUpgradeRequired              = problems.ProblemUpgradeRequired
	ProblemPreconditionRequired         = problems.ProblemPreconditionRequired
	ProblemTooManyRequests              = problems.ProblemTooManyRequests
	ProblemRequestHeaderFieldsTooLarge  = problems.ProblemRequestHeaderFieldsTooLarge
	ProblemUnavailableForLegalReasons   = problems.ProblemUnavailableForLegalReasons

	// 500
	ProblemInternalServerError           = problems.ProblemInternalServerError
	ProblemNotImplemented                = problems.ProblemNotImplemented
	ProblemBadGateway                    = problems.ProblemBadGateway
	ProblemServiceUnavailable            = problems.ProblemServiceUnavailable
	ProblemGatewayTimeout                = problems.ProblemGatewayTimeout
	ProblemHTTPVersionNotSupported       = problems.ProblemHTTPVersionNotSupported
	ProblemVariantAlsoNegotiates         = problems.ProblemVariantAlsoNegotiates
	ProblemInsufficientStorage           = problems.ProblemInsufficientStorage
	ProblemLoopDetected                  = problems.ProblemLoopDetected
	ProblemNotExtended                   = problems.ProblemNotExtended
	ProblemNetworkAuthenticationRequired = problems.ProblemNetworkAuthenticationRequired
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
	CTXFilter      = server.CTXFilter
	FilterProblem  = server.FilterProblem
	Cache          = cache.Cache
	Logger         = logs.Logger
	Logs           = logs.Logs

	// FieldError 表示配置文件中的字段错误
	FieldError = config.FieldError

	// QueryUnmarshaler 对查询参数的解析接口
	QueryUnmarshaler = query.Unmarshaler

	// LocaleStringer 本地化字符串需要实在的接口
	LocaleStringer = localeutil.LocaleStringer

	StringPhrase = localeutil.StringPhrase
)

func NewServer(name, version string, o *Options) (*Server, error) {
	return server.New(name, version, o)
}

// Phrase 生成本地化的语言片段
func Phrase(key string, v ...any) LocaleStringer {
	return localeutil.Phrase(key, v...)
}

// NewStackError 为 err 带上调用信息
//
// 位置从调用 NewStackError 开始。如果 err 为 nil，则返回 nil。
// 多次调用 NewStackError 包装，则返回第一次包装的返回值。
//
// 如果需要输出调用堆栈信息，需要指定 %+v 标记。
func NewStackError(err error) error { return errs.NewDepthStackError(2, err) }

// NewFieldError 返回表示配置文件错误的对象
//
// field 表示错误的字段名；
// msg 表示错误信息，可以是任意类型，如果 msg 是 [FieldError] 类型，
// 那么此操作相当于调用了 [FieldError.AddFieldParent]；
func NewFieldError(field string, msg any) *FieldError {
	return config.NewFieldError(field, msg)
}

// NewLocaleError 本地化的错误信息
func NewLocaleError(format string, v ...any) error {
	return localeutil.Error(format, v...)
}

// NewHTTPError 用 HTTP 状态码包装一个错误信息
//
// 此方法返回的错误，在 [Context.Error] 会被识别且按指定的状态码输出。
func NewHTTPError(status int, err error) error {
	return errs.NewHTTPError(status, err)
}
