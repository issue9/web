// 此文件由工具产生，请勿手动修改！

package web

import (
	"net/http"

	"github.com/issue9/web/internal/problems"
)

const (
	ProblemAboutBlank = problems.AboutBlank

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

var problemsID = map[int]string{
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

func initProblems(p *problems.Problems) {
	p.Add(ProblemBadRequest, http.StatusBadRequest, StringPhrase("problem.400"), StringPhrase("problem.400.detail"))
	p.Add(ProblemUnauthorized, http.StatusUnauthorized, StringPhrase("problem.401"), StringPhrase("problem.401.detail"))
	p.Add(ProblemPaymentRequired, http.StatusPaymentRequired, StringPhrase("problem.402"), StringPhrase("problem.402.detail"))
	p.Add(ProblemForbidden, http.StatusForbidden, StringPhrase("problem.403"), StringPhrase("problem.403.detail"))
	p.Add(ProblemNotFound, http.StatusNotFound, StringPhrase("problem.404"), StringPhrase("problem.404.detail"))
	p.Add(ProblemMethodNotAllowed, http.StatusMethodNotAllowed, StringPhrase("problem.405"), StringPhrase("problem.405.detail"))
	p.Add(ProblemNotAcceptable, http.StatusNotAcceptable, StringPhrase("problem.406"), StringPhrase("problem.406.detail"))
	p.Add(ProblemProxyAuthRequired, http.StatusProxyAuthRequired, StringPhrase("problem.407"), StringPhrase("problem.407.detail"))
	p.Add(ProblemRequestTimeout, http.StatusRequestTimeout, StringPhrase("problem.408"), StringPhrase("problem.408.detail"))
	p.Add(ProblemConflict, http.StatusConflict, StringPhrase("problem.409"), StringPhrase("problem.409.detail"))
	p.Add(ProblemGone, http.StatusGone, StringPhrase("problem.410"), StringPhrase("problem.410.detail"))
	p.Add(ProblemLengthRequired, http.StatusLengthRequired, StringPhrase("problem.411"), StringPhrase("problem.411.detail"))
	p.Add(ProblemPreconditionFailed, http.StatusPreconditionFailed, StringPhrase("problem.412"), StringPhrase("problem.412.detail"))
	p.Add(ProblemRequestEntityTooLarge, http.StatusRequestEntityTooLarge, StringPhrase("problem.413"), StringPhrase("problem.413.detail"))
	p.Add(ProblemRequestURITooLong, http.StatusRequestURITooLong, StringPhrase("problem.414"), StringPhrase("problem.414.detail"))
	p.Add(ProblemUnsupportedMediaType, http.StatusUnsupportedMediaType, StringPhrase("problem.415"), StringPhrase("problem.415.detail"))
	p.Add(ProblemRequestedRangeNotSatisfiable, http.StatusRequestedRangeNotSatisfiable, StringPhrase("problem.416"), StringPhrase("problem.416.detail"))
	p.Add(ProblemExpectationFailed, http.StatusExpectationFailed, StringPhrase("problem.417"), StringPhrase("problem.417.detail"))
	p.Add(ProblemTeapot, http.StatusTeapot, StringPhrase("problem.418"), StringPhrase("problem.418.detail"))
	p.Add(ProblemMisdirectedRequest, http.StatusMisdirectedRequest, StringPhrase("problem.421"), StringPhrase("problem.421.detail"))
	p.Add(ProblemUnprocessableEntity, http.StatusUnprocessableEntity, StringPhrase("problem.422"), StringPhrase("problem.422.detail"))
	p.Add(ProblemLocked, http.StatusLocked, StringPhrase("problem.423"), StringPhrase("problem.423.detail"))
	p.Add(ProblemFailedDependency, http.StatusFailedDependency, StringPhrase("problem.424"), StringPhrase("problem.424.detail"))
	p.Add(ProblemTooEarly, http.StatusTooEarly, StringPhrase("problem.425"), StringPhrase("problem.425.detail"))
	p.Add(ProblemUpgradeRequired, http.StatusUpgradeRequired, StringPhrase("problem.426"), StringPhrase("problem.426.detail"))
	p.Add(ProblemPreconditionRequired, http.StatusPreconditionRequired, StringPhrase("problem.428"), StringPhrase("problem.428.detail"))
	p.Add(ProblemTooManyRequests, http.StatusTooManyRequests, StringPhrase("problem.429"), StringPhrase("problem.429.detail"))
	p.Add(ProblemRequestHeaderFieldsTooLarge, http.StatusRequestHeaderFieldsTooLarge, StringPhrase("problem.431"), StringPhrase("problem.431.detail"))
	p.Add(ProblemUnavailableForLegalReasons, http.StatusUnavailableForLegalReasons, StringPhrase("problem.451"), StringPhrase("problem.451.detail"))
	p.Add(ProblemInternalServerError, http.StatusInternalServerError, StringPhrase("problem.500"), StringPhrase("problem.500.detail"))
	p.Add(ProblemNotImplemented, http.StatusNotImplemented, StringPhrase("problem.501"), StringPhrase("problem.501.detail"))
	p.Add(ProblemBadGateway, http.StatusBadGateway, StringPhrase("problem.502"), StringPhrase("problem.502.detail"))
	p.Add(ProblemServiceUnavailable, http.StatusServiceUnavailable, StringPhrase("problem.503"), StringPhrase("problem.503.detail"))
	p.Add(ProblemGatewayTimeout, http.StatusGatewayTimeout, StringPhrase("problem.504"), StringPhrase("problem.504.detail"))
	p.Add(ProblemHTTPVersionNotSupported, http.StatusHTTPVersionNotSupported, StringPhrase("problem.505"), StringPhrase("problem.505.detail"))
	p.Add(ProblemVariantAlsoNegotiates, http.StatusVariantAlsoNegotiates, StringPhrase("problem.506"), StringPhrase("problem.506.detail"))
	p.Add(ProblemInsufficientStorage, http.StatusInsufficientStorage, StringPhrase("problem.507"), StringPhrase("problem.507.detail"))
	p.Add(ProblemLoopDetected, http.StatusLoopDetected, StringPhrase("problem.508"), StringPhrase("problem.508.detail"))
	p.Add(ProblemNotExtended, http.StatusNotExtended, StringPhrase("problem.510"), StringPhrase("problem.510.detail"))
	p.Add(ProblemNetworkAuthenticationRequired, http.StatusNetworkAuthenticationRequired, StringPhrase("problem.511"), StringPhrase("problem.511.detail"))
}
