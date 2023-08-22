// 此文件由工具产生，请勿手动修改！

package problems

import (
	"net/http"

	"github.com/issue9/localeutil"
)

const (
	ProblemAboutBlank = "about:blank"

	ProblemBadRequest                    = "400"
	ProblemUnauthorized                  = "401"
	ProblemPaymentRequired               = "402"
	ProblemForbidden                     = "403"
	ProblemNotFound                      = "404"
	ProblemMethodNotAllowed              = "405"
	ProblemNotAcceptable                 = "406"
	ProblemProxyAuthRequired             = "407"
	ProblemRequestTimeout                = "408"
	ProblemConflict                      = "409"
	ProblemGone                          = "410"
	ProblemLengthRequired                = "411"
	ProblemPreconditionFailed            = "412"
	ProblemRequestEntityTooLarge         = "413"
	ProblemRequestURITooLong             = "414"
	ProblemUnsupportedMediaType          = "415"
	ProblemRequestedRangeNotSatisfiable  = "416"
	ProblemExpectationFailed             = "417"
	ProblemTeapot                        = "418"
	ProblemMisdirectedRequest            = "421"
	ProblemUnprocessableEntity           = "422"
	ProblemLocked                        = "423"
	ProblemFailedDependency              = "424"
	ProblemTooEarly                      = "425"
	ProblemUpgradeRequired               = "426"
	ProblemPreconditionRequired          = "428"
	ProblemTooManyRequests               = "429"
	ProblemRequestHeaderFieldsTooLarge   = "431"
	ProblemUnavailableForLegalReasons    = "451"
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
	ProblemBadRequest:                    http.StatusBadRequest,
	ProblemUnauthorized:                  http.StatusUnauthorized,
	ProblemPaymentRequired:               http.StatusPaymentRequired,
	ProblemForbidden:                     http.StatusForbidden,
	ProblemNotFound:                      http.StatusNotFound,
	ProblemMethodNotAllowed:              http.StatusMethodNotAllowed,
	ProblemNotAcceptable:                 http.StatusNotAcceptable,
	ProblemProxyAuthRequired:             http.StatusProxyAuthRequired,
	ProblemRequestTimeout:                http.StatusRequestTimeout,
	ProblemConflict:                      http.StatusConflict,
	ProblemGone:                          http.StatusGone,
	ProblemLengthRequired:                http.StatusLengthRequired,
	ProblemPreconditionFailed:            http.StatusPreconditionFailed,
	ProblemRequestEntityTooLarge:         http.StatusRequestEntityTooLarge,
	ProblemRequestURITooLong:             http.StatusRequestURITooLong,
	ProblemUnsupportedMediaType:          http.StatusUnsupportedMediaType,
	ProblemRequestedRangeNotSatisfiable:  http.StatusRequestedRangeNotSatisfiable,
	ProblemExpectationFailed:             http.StatusExpectationFailed,
	ProblemTeapot:                        http.StatusTeapot,
	ProblemMisdirectedRequest:            http.StatusMisdirectedRequest,
	ProblemUnprocessableEntity:           http.StatusUnprocessableEntity,
	ProblemLocked:                        http.StatusLocked,
	ProblemFailedDependency:              http.StatusFailedDependency,
	ProblemTooEarly:                      http.StatusTooEarly,
	ProblemUpgradeRequired:               http.StatusUpgradeRequired,
	ProblemPreconditionRequired:          http.StatusPreconditionRequired,
	ProblemTooManyRequests:               http.StatusTooManyRequests,
	ProblemRequestHeaderFieldsTooLarge:   http.StatusRequestHeaderFieldsTooLarge,
	ProblemUnavailableForLegalReasons:    http.StatusUnavailableForLegalReasons,
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

var ids = map[int]string{
	http.StatusBadRequest:                    ProblemBadRequest,
	http.StatusUnauthorized:                  ProblemUnauthorized,
	http.StatusPaymentRequired:               ProblemPaymentRequired,
	http.StatusForbidden:                     ProblemForbidden,
	http.StatusNotFound:                      ProblemNotFound,
	http.StatusMethodNotAllowed:              ProblemMethodNotAllowed,
	http.StatusNotAcceptable:                 ProblemNotAcceptable,
	http.StatusProxyAuthRequired:             ProblemProxyAuthRequired,
	http.StatusRequestTimeout:                ProblemRequestTimeout,
	http.StatusConflict:                      ProblemConflict,
	http.StatusGone:                          ProblemGone,
	http.StatusLengthRequired:                ProblemLengthRequired,
	http.StatusPreconditionFailed:            ProblemPreconditionFailed,
	http.StatusRequestEntityTooLarge:         ProblemRequestEntityTooLarge,
	http.StatusRequestURITooLong:             ProblemRequestURITooLong,
	http.StatusUnsupportedMediaType:          ProblemUnsupportedMediaType,
	http.StatusRequestedRangeNotSatisfiable:  ProblemRequestedRangeNotSatisfiable,
	http.StatusExpectationFailed:             ProblemExpectationFailed,
	http.StatusTeapot:                        ProblemTeapot,
	http.StatusMisdirectedRequest:            ProblemMisdirectedRequest,
	http.StatusUnprocessableEntity:           ProblemUnprocessableEntity,
	http.StatusLocked:                        ProblemLocked,
	http.StatusFailedDependency:              ProblemFailedDependency,
	http.StatusTooEarly:                      ProblemTooEarly,
	http.StatusUpgradeRequired:               ProblemUpgradeRequired,
	http.StatusPreconditionRequired:          ProblemPreconditionRequired,
	http.StatusTooManyRequests:               ProblemTooManyRequests,
	http.StatusRequestHeaderFieldsTooLarge:   ProblemRequestHeaderFieldsTooLarge,
	http.StatusUnavailableForLegalReasons:    ProblemUnavailableForLegalReasons,
	http.StatusInternalServerError:           ProblemInternalServerError,
	http.StatusNotImplemented:                ProblemNotImplemented,
	http.StatusBadGateway:                    ProblemBadGateway,
	http.StatusServiceUnavailable:            ProblemServiceUnavailable,
	http.StatusGatewayTimeout:                ProblemGatewayTimeout,
	http.StatusHTTPVersionNotSupported:       ProblemHTTPVersionNotSupported,
	http.StatusVariantAlsoNegotiates:         ProblemVariantAlsoNegotiates,
	http.StatusInsufficientStorage:           ProblemInsufficientStorage,
	http.StatusLoopDetected:                  ProblemLoopDetected,
	http.StatusNotExtended:                   ProblemNotExtended,
	http.StatusNetworkAuthenticationRequired: ProblemNetworkAuthenticationRequired,
}

func (p *Problems) initLocales() {
	p.Add(ProblemBadRequest, http.StatusBadRequest, localeutil.StringPhrase("problem.400"), localeutil.StringPhrase("problem.400.detail"))
	p.Add(ProblemUnauthorized, http.StatusUnauthorized, localeutil.StringPhrase("problem.401"), localeutil.StringPhrase("problem.401.detail"))
	p.Add(ProblemPaymentRequired, http.StatusPaymentRequired, localeutil.StringPhrase("problem.402"), localeutil.StringPhrase("problem.402.detail"))
	p.Add(ProblemForbidden, http.StatusForbidden, localeutil.StringPhrase("problem.403"), localeutil.StringPhrase("problem.403.detail"))
	p.Add(ProblemNotFound, http.StatusNotFound, localeutil.StringPhrase("problem.404"), localeutil.StringPhrase("problem.404.detail"))
	p.Add(ProblemMethodNotAllowed, http.StatusMethodNotAllowed, localeutil.StringPhrase("problem.405"), localeutil.StringPhrase("problem.405.detail"))
	p.Add(ProblemNotAcceptable, http.StatusNotAcceptable, localeutil.StringPhrase("problem.406"), localeutil.StringPhrase("problem.406.detail"))
	p.Add(ProblemProxyAuthRequired, http.StatusProxyAuthRequired, localeutil.StringPhrase("problem.407"), localeutil.StringPhrase("problem.407.detail"))
	p.Add(ProblemRequestTimeout, http.StatusRequestTimeout, localeutil.StringPhrase("problem.408"), localeutil.StringPhrase("problem.408.detail"))
	p.Add(ProblemConflict, http.StatusConflict, localeutil.StringPhrase("problem.409"), localeutil.StringPhrase("problem.409.detail"))
	p.Add(ProblemGone, http.StatusGone, localeutil.StringPhrase("problem.410"), localeutil.StringPhrase("problem.410.detail"))
	p.Add(ProblemLengthRequired, http.StatusLengthRequired, localeutil.StringPhrase("problem.411"), localeutil.StringPhrase("problem.411.detail"))
	p.Add(ProblemPreconditionFailed, http.StatusPreconditionFailed, localeutil.StringPhrase("problem.412"), localeutil.StringPhrase("problem.412.detail"))
	p.Add(ProblemRequestEntityTooLarge, http.StatusRequestEntityTooLarge, localeutil.StringPhrase("problem.413"), localeutil.StringPhrase("problem.413.detail"))
	p.Add(ProblemRequestURITooLong, http.StatusRequestURITooLong, localeutil.StringPhrase("problem.414"), localeutil.StringPhrase("problem.414.detail"))
	p.Add(ProblemUnsupportedMediaType, http.StatusUnsupportedMediaType, localeutil.StringPhrase("problem.415"), localeutil.StringPhrase("problem.415.detail"))
	p.Add(ProblemRequestedRangeNotSatisfiable, http.StatusRequestedRangeNotSatisfiable, localeutil.StringPhrase("problem.416"), localeutil.StringPhrase("problem.416.detail"))
	p.Add(ProblemExpectationFailed, http.StatusExpectationFailed, localeutil.StringPhrase("problem.417"), localeutil.StringPhrase("problem.417.detail"))
	p.Add(ProblemTeapot, http.StatusTeapot, localeutil.StringPhrase("problem.418"), localeutil.StringPhrase("problem.418.detail"))
	p.Add(ProblemMisdirectedRequest, http.StatusMisdirectedRequest, localeutil.StringPhrase("problem.421"), localeutil.StringPhrase("problem.421.detail"))
	p.Add(ProblemUnprocessableEntity, http.StatusUnprocessableEntity, localeutil.StringPhrase("problem.422"), localeutil.StringPhrase("problem.422.detail"))
	p.Add(ProblemLocked, http.StatusLocked, localeutil.StringPhrase("problem.423"), localeutil.StringPhrase("problem.423.detail"))
	p.Add(ProblemFailedDependency, http.StatusFailedDependency, localeutil.StringPhrase("problem.424"), localeutil.StringPhrase("problem.424.detail"))
	p.Add(ProblemTooEarly, http.StatusTooEarly, localeutil.StringPhrase("problem.425"), localeutil.StringPhrase("problem.425.detail"))
	p.Add(ProblemUpgradeRequired, http.StatusUpgradeRequired, localeutil.StringPhrase("problem.426"), localeutil.StringPhrase("problem.426.detail"))
	p.Add(ProblemPreconditionRequired, http.StatusPreconditionRequired, localeutil.StringPhrase("problem.428"), localeutil.StringPhrase("problem.428.detail"))
	p.Add(ProblemTooManyRequests, http.StatusTooManyRequests, localeutil.StringPhrase("problem.429"), localeutil.StringPhrase("problem.429.detail"))
	p.Add(ProblemRequestHeaderFieldsTooLarge, http.StatusRequestHeaderFieldsTooLarge, localeutil.StringPhrase("problem.431"), localeutil.StringPhrase("problem.431.detail"))
	p.Add(ProblemUnavailableForLegalReasons, http.StatusUnavailableForLegalReasons, localeutil.StringPhrase("problem.451"), localeutil.StringPhrase("problem.451.detail"))
	p.Add(ProblemInternalServerError, http.StatusInternalServerError, localeutil.StringPhrase("problem.500"), localeutil.StringPhrase("problem.500.detail"))
	p.Add(ProblemNotImplemented, http.StatusNotImplemented, localeutil.StringPhrase("problem.501"), localeutil.StringPhrase("problem.501.detail"))
	p.Add(ProblemBadGateway, http.StatusBadGateway, localeutil.StringPhrase("problem.502"), localeutil.StringPhrase("problem.502.detail"))
	p.Add(ProblemServiceUnavailable, http.StatusServiceUnavailable, localeutil.StringPhrase("problem.503"), localeutil.StringPhrase("problem.503.detail"))
	p.Add(ProblemGatewayTimeout, http.StatusGatewayTimeout, localeutil.StringPhrase("problem.504"), localeutil.StringPhrase("problem.504.detail"))
	p.Add(ProblemHTTPVersionNotSupported, http.StatusHTTPVersionNotSupported, localeutil.StringPhrase("problem.505"), localeutil.StringPhrase("problem.505.detail"))
	p.Add(ProblemVariantAlsoNegotiates, http.StatusVariantAlsoNegotiates, localeutil.StringPhrase("problem.506"), localeutil.StringPhrase("problem.506.detail"))
	p.Add(ProblemInsufficientStorage, http.StatusInsufficientStorage, localeutil.StringPhrase("problem.507"), localeutil.StringPhrase("problem.507.detail"))
	p.Add(ProblemLoopDetected, http.StatusLoopDetected, localeutil.StringPhrase("problem.508"), localeutil.StringPhrase("problem.508.detail"))
	p.Add(ProblemNotExtended, http.StatusNotExtended, localeutil.StringPhrase("problem.510"), localeutil.StringPhrase("problem.510.detail"))
	p.Add(ProblemNetworkAuthenticationRequired, http.StatusNetworkAuthenticationRequired, localeutil.StringPhrase("problem.511"), localeutil.StringPhrase("problem.511.detail"))
}
