// 此文件由工具产生，请勿手动修改！

package web

import "net/http"

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

func initProblems(p *Problems) {
	p.Add(http.StatusBadRequest, &LocaleProblem{ID: ProblemBadRequest, Title: StringPhrase("problem.400"), Detail: StringPhrase("problem.400.detail")})
	p.Add(http.StatusUnauthorized, &LocaleProblem{ID: ProblemUnauthorized, Title: StringPhrase("problem.401"), Detail: StringPhrase("problem.401.detail")})
	p.Add(http.StatusPaymentRequired, &LocaleProblem{ID: ProblemPaymentRequired, Title: StringPhrase("problem.402"), Detail: StringPhrase("problem.402.detail")})
	p.Add(http.StatusForbidden, &LocaleProblem{ID: ProblemForbidden, Title: StringPhrase("problem.403"), Detail: StringPhrase("problem.403.detail")})
	p.Add(http.StatusNotFound, &LocaleProblem{ID: ProblemNotFound, Title: StringPhrase("problem.404"), Detail: StringPhrase("problem.404.detail")})
	p.Add(http.StatusMethodNotAllowed, &LocaleProblem{ID: ProblemMethodNotAllowed, Title: StringPhrase("problem.405"), Detail: StringPhrase("problem.405.detail")})
	p.Add(http.StatusNotAcceptable, &LocaleProblem{ID: ProblemNotAcceptable, Title: StringPhrase("problem.406"), Detail: StringPhrase("problem.406.detail")})
	p.Add(http.StatusProxyAuthRequired, &LocaleProblem{ID: ProblemProxyAuthRequired, Title: StringPhrase("problem.407"), Detail: StringPhrase("problem.407.detail")})
	p.Add(http.StatusRequestTimeout, &LocaleProblem{ID: ProblemRequestTimeout, Title: StringPhrase("problem.408"), Detail: StringPhrase("problem.408.detail")})
	p.Add(http.StatusConflict, &LocaleProblem{ID: ProblemConflict, Title: StringPhrase("problem.409"), Detail: StringPhrase("problem.409.detail")})
	p.Add(http.StatusGone, &LocaleProblem{ID: ProblemGone, Title: StringPhrase("problem.410"), Detail: StringPhrase("problem.410.detail")})
	p.Add(http.StatusLengthRequired, &LocaleProblem{ID: ProblemLengthRequired, Title: StringPhrase("problem.411"), Detail: StringPhrase("problem.411.detail")})
	p.Add(http.StatusPreconditionFailed, &LocaleProblem{ID: ProblemPreconditionFailed, Title: StringPhrase("problem.412"), Detail: StringPhrase("problem.412.detail")})
	p.Add(http.StatusRequestEntityTooLarge, &LocaleProblem{ID: ProblemRequestEntityTooLarge, Title: StringPhrase("problem.413"), Detail: StringPhrase("problem.413.detail")})
	p.Add(http.StatusRequestURITooLong, &LocaleProblem{ID: ProblemRequestURITooLong, Title: StringPhrase("problem.414"), Detail: StringPhrase("problem.414.detail")})
	p.Add(http.StatusUnsupportedMediaType, &LocaleProblem{ID: ProblemUnsupportedMediaType, Title: StringPhrase("problem.415"), Detail: StringPhrase("problem.415.detail")})
	p.Add(http.StatusRequestedRangeNotSatisfiable, &LocaleProblem{ID: ProblemRequestedRangeNotSatisfiable, Title: StringPhrase("problem.416"), Detail: StringPhrase("problem.416.detail")})
	p.Add(http.StatusExpectationFailed, &LocaleProblem{ID: ProblemExpectationFailed, Title: StringPhrase("problem.417"), Detail: StringPhrase("problem.417.detail")})
	p.Add(http.StatusTeapot, &LocaleProblem{ID: ProblemTeapot, Title: StringPhrase("problem.418"), Detail: StringPhrase("problem.418.detail")})
	p.Add(http.StatusMisdirectedRequest, &LocaleProblem{ID: ProblemMisdirectedRequest, Title: StringPhrase("problem.421"), Detail: StringPhrase("problem.421.detail")})
	p.Add(http.StatusUnprocessableEntity, &LocaleProblem{ID: ProblemUnprocessableEntity, Title: StringPhrase("problem.422"), Detail: StringPhrase("problem.422.detail")})
	p.Add(http.StatusLocked, &LocaleProblem{ID: ProblemLocked, Title: StringPhrase("problem.423"), Detail: StringPhrase("problem.423.detail")})
	p.Add(http.StatusFailedDependency, &LocaleProblem{ID: ProblemFailedDependency, Title: StringPhrase("problem.424"), Detail: StringPhrase("problem.424.detail")})
	p.Add(http.StatusTooEarly, &LocaleProblem{ID: ProblemTooEarly, Title: StringPhrase("problem.425"), Detail: StringPhrase("problem.425.detail")})
	p.Add(http.StatusUpgradeRequired, &LocaleProblem{ID: ProblemUpgradeRequired, Title: StringPhrase("problem.426"), Detail: StringPhrase("problem.426.detail")})
	p.Add(http.StatusPreconditionRequired, &LocaleProblem{ID: ProblemPreconditionRequired, Title: StringPhrase("problem.428"), Detail: StringPhrase("problem.428.detail")})
	p.Add(http.StatusTooManyRequests, &LocaleProblem{ID: ProblemTooManyRequests, Title: StringPhrase("problem.429"), Detail: StringPhrase("problem.429.detail")})
	p.Add(http.StatusRequestHeaderFieldsTooLarge, &LocaleProblem{ID: ProblemRequestHeaderFieldsTooLarge, Title: StringPhrase("problem.431"), Detail: StringPhrase("problem.431.detail")})
	p.Add(http.StatusUnavailableForLegalReasons, &LocaleProblem{ID: ProblemUnavailableForLegalReasons, Title: StringPhrase("problem.451"), Detail: StringPhrase("problem.451.detail")})
	p.Add(http.StatusInternalServerError, &LocaleProblem{ID: ProblemInternalServerError, Title: StringPhrase("problem.500"), Detail: StringPhrase("problem.500.detail")})
	p.Add(http.StatusNotImplemented, &LocaleProblem{ID: ProblemNotImplemented, Title: StringPhrase("problem.501"), Detail: StringPhrase("problem.501.detail")})
	p.Add(http.StatusBadGateway, &LocaleProblem{ID: ProblemBadGateway, Title: StringPhrase("problem.502"), Detail: StringPhrase("problem.502.detail")})
	p.Add(http.StatusServiceUnavailable, &LocaleProblem{ID: ProblemServiceUnavailable, Title: StringPhrase("problem.503"), Detail: StringPhrase("problem.503.detail")})
	p.Add(http.StatusGatewayTimeout, &LocaleProblem{ID: ProblemGatewayTimeout, Title: StringPhrase("problem.504"), Detail: StringPhrase("problem.504.detail")})
	p.Add(http.StatusHTTPVersionNotSupported, &LocaleProblem{ID: ProblemHTTPVersionNotSupported, Title: StringPhrase("problem.505"), Detail: StringPhrase("problem.505.detail")})
	p.Add(http.StatusVariantAlsoNegotiates, &LocaleProblem{ID: ProblemVariantAlsoNegotiates, Title: StringPhrase("problem.506"), Detail: StringPhrase("problem.506.detail")})
	p.Add(http.StatusInsufficientStorage, &LocaleProblem{ID: ProblemInsufficientStorage, Title: StringPhrase("problem.507"), Detail: StringPhrase("problem.507.detail")})
	p.Add(http.StatusLoopDetected, &LocaleProblem{ID: ProblemLoopDetected, Title: StringPhrase("problem.508"), Detail: StringPhrase("problem.508.detail")})
	p.Add(http.StatusNotExtended, &LocaleProblem{ID: ProblemNotExtended, Title: StringPhrase("problem.510"), Detail: StringPhrase("problem.510.detail")})
	p.Add(http.StatusNetworkAuthenticationRequired, &LocaleProblem{ID: ProblemNetworkAuthenticationRequired, Title: StringPhrase("problem.511"), Detail: StringPhrase("problem.511.detail")})
}
