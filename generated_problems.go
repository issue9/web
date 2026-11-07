// 此文件由工具产生，请勿手动修改！

package web

import "net/http"

const (
	aboutBlank = "about:blank"

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
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemBadRequest, Title: StringPhrase("problem.400"), Status: http.StatusBadRequest})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemUnauthorized, Title: StringPhrase("problem.401"), Status: http.StatusUnauthorized})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemPaymentRequired, Title: StringPhrase("problem.402"), Status: http.StatusPaymentRequired})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemForbidden, Title: StringPhrase("problem.403"), Status: http.StatusForbidden})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemNotFound, Title: StringPhrase("problem.404"), Status: http.StatusNotFound})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemMethodNotAllowed, Title: StringPhrase("problem.405"), Status: http.StatusMethodNotAllowed})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemNotAcceptable, Title: StringPhrase("problem.406"), Status: http.StatusNotAcceptable})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemProxyAuthRequired, Title: StringPhrase("problem.407"), Status: http.StatusProxyAuthRequired})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemRequestTimeout, Title: StringPhrase("problem.408"), Status: http.StatusRequestTimeout})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemConflict, Title: StringPhrase("problem.409"), Status: http.StatusConflict})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemGone, Title: StringPhrase("problem.410"), Status: http.StatusGone})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemLengthRequired, Title: StringPhrase("problem.411"), Status: http.StatusLengthRequired})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemPreconditionFailed, Title: StringPhrase("problem.412"), Status: http.StatusPreconditionFailed})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemRequestEntityTooLarge, Title: StringPhrase("problem.413"), Status: http.StatusRequestEntityTooLarge})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemRequestURITooLong, Title: StringPhrase("problem.414"), Status: http.StatusRequestURITooLong})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemUnsupportedMediaType, Title: StringPhrase("problem.415"), Status: http.StatusUnsupportedMediaType})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemRequestedRangeNotSatisfiable, Title: StringPhrase("problem.416"), Status: http.StatusRequestedRangeNotSatisfiable})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemExpectationFailed, Title: StringPhrase("problem.417"), Status: http.StatusExpectationFailed})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemTeapot, Title: StringPhrase("problem.418"), Status: http.StatusTeapot})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemMisdirectedRequest, Title: StringPhrase("problem.421"), Status: http.StatusMisdirectedRequest})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemUnprocessableEntity, Title: StringPhrase("problem.422"), Status: http.StatusUnprocessableEntity})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemLocked, Title: StringPhrase("problem.423"), Status: http.StatusLocked})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemFailedDependency, Title: StringPhrase("problem.424"), Status: http.StatusFailedDependency})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemTooEarly, Title: StringPhrase("problem.425"), Status: http.StatusTooEarly})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemUpgradeRequired, Title: StringPhrase("problem.426"), Status: http.StatusUpgradeRequired})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemPreconditionRequired, Title: StringPhrase("problem.428"), Status: http.StatusPreconditionRequired})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemTooManyRequests, Title: StringPhrase("problem.429"), Status: http.StatusTooManyRequests})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemRequestHeaderFieldsTooLarge, Title: StringPhrase("problem.431"), Status: http.StatusRequestHeaderFieldsTooLarge})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemUnavailableForLegalReasons, Title: StringPhrase("problem.451"), Status: http.StatusUnavailableForLegalReasons})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemInternalServerError, Title: StringPhrase("problem.500"), Status: http.StatusInternalServerError})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemNotImplemented, Title: StringPhrase("problem.501"), Status: http.StatusNotImplemented})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemBadGateway, Title: StringPhrase("problem.502"), Status: http.StatusBadGateway})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemServiceUnavailable, Title: StringPhrase("problem.503"), Status: http.StatusServiceUnavailable})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemGatewayTimeout, Title: StringPhrase("problem.504"), Status: http.StatusGatewayTimeout})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemHTTPVersionNotSupported, Title: StringPhrase("problem.505"), Status: http.StatusHTTPVersionNotSupported})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemVariantAlsoNegotiates, Title: StringPhrase("problem.506"), Status: http.StatusVariantAlsoNegotiates})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemInsufficientStorage, Title: StringPhrase("problem.507"), Status: http.StatusInsufficientStorage})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemLoopDetected, Title: StringPhrase("problem.508"), Status: http.StatusLoopDetected})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemNotExtended, Title: StringPhrase("problem.510"), Status: http.StatusNotExtended})
	p.add(&LocaleProblem{Type: "about:blank", ID: ProblemNetworkAuthenticationRequired, Title: StringPhrase("problem.511"), Status: http.StatusNetworkAuthenticationRequired})
}
