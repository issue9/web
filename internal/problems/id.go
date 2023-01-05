// SPDX-License-Identifier: MIT

package problems

import "net/http"

// 预定义的 Problem id 值
const (
	// 特殊值，当不希望显示 type 值时，将 type 赋予此值，不应该出现相应的状态码。
	ProblemAboutBlank = "about:blank"

	// NOTE: 如果以下添加或是删除了值，应该同时处理 statuses，以及 /web.go 中的引用！

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

var statuses = map[string]int{
	ProblemBadRequest:                   http.StatusBadRequest,
	ProblemUnauthorized:                 http.StatusUnauthorized,
	ProblemPaymentRequired:              http.StatusPaymentRequired,
	ProblemForbidden:                    http.StatusForbidden,
	ProblemNotFound:                     http.StatusNotFound,
	ProblemMethodNotAllowed:             http.StatusMethodNotAllowed,
	ProblemNotAcceptable:                http.StatusNotAcceptable,
	ProblemProxyAuthRequired:            http.StatusProxyAuthRequired,
	ProblemRequestTimeout:               http.StatusRequestTimeout,
	ProblemConflict:                     http.StatusConflict,
	ProblemGone:                         http.StatusGone,
	ProblemLengthRequired:               http.StatusLengthRequired,
	ProblemPreconditionFailed:           http.StatusPreconditionFailed,
	ProblemRequestEntityTooLarge:        http.StatusRequestEntityTooLarge,
	ProblemRequestURITooLong:            http.StatusRequestURITooLong,
	ProblemUnsupportedMediaType:         http.StatusUnsupportedMediaType,
	ProblemRequestedRangeNotSatisfiable: http.StatusRequestedRangeNotSatisfiable,
	ProblemExpectationFailed:            http.StatusExpectationFailed,
	ProblemTeapot:                       http.StatusTeapot,
	ProblemMisdirectedRequest:           http.StatusMisdirectedRequest,
	ProblemUnprocessableEntity:          http.StatusUnprocessableEntity,
	ProblemLocked:                       http.StatusLocked,
	ProblemFailedDependency:             http.StatusFailedDependency,
	ProblemTooEarly:                     http.StatusTooEarly,
	ProblemUpgradeRequired:              http.StatusUpgradeRequired,
	ProblemPreconditionRequired:         http.StatusPreconditionRequired,
	ProblemTooManyRequests:              http.StatusTooManyRequests,
	ProblemRequestHeaderFieldsTooLarge:  http.StatusRequestHeaderFieldsTooLarge,
	ProblemUnavailableForLegalReasons:   http.StatusUnavailableForLegalReasons,

	// 500
	ProblemInternalServerError:           http.StatusInternalServerError,
	ProblemNotImplemented:                http.StatusNotImplemented,
	ProblemBadGateway:                    http.StatusBadGateway,
	ProblemServiceUnavailable:            http.StatusServiceUnavailable,
	ProblemGatewayTimeout:                http.StatusGatewayTimeout,
	ProblemHTTPVersionNotSupported:       http.StatusHTTPVersionNotSupported,
	ProblemVariantAlsoNegotiates:         http.StatusVariantAlsoNegotiates,
	ProblemInsufficientStorage:           http.StatusInsufficientStorage,
	ProblemLoopDetected:                  http.StatusLoopDetected,
	ProblemNotExtended:                   http.StatusNotExtended,
	ProblemNetworkAuthenticationRequired: http.StatusNetworkAuthenticationRequired,
}

func Status(id string) int {
	return statuses[id]
}
