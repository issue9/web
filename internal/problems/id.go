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

var locales = map[int]localeutil.LocaleStringer{
	http.StatusBadRequest:                    localeutil.StringPhrase("problem.400"),
	http.StatusUnauthorized:                  localeutil.StringPhrase("problem.401"),
	http.StatusPaymentRequired:               localeutil.StringPhrase("problem.402"),
	http.StatusForbidden:                     localeutil.StringPhrase("problem.403"),
	http.StatusNotFound:                      localeutil.StringPhrase("problem.404"),
	http.StatusMethodNotAllowed:              localeutil.StringPhrase("problem.405"),
	http.StatusNotAcceptable:                 localeutil.StringPhrase("problem.406"),
	http.StatusProxyAuthRequired:             localeutil.StringPhrase("problem.407"),
	http.StatusRequestTimeout:                localeutil.StringPhrase("problem.408"),
	http.StatusConflict:                      localeutil.StringPhrase("problem.409"),
	http.StatusGone:                          localeutil.StringPhrase("problem.410"),
	http.StatusLengthRequired:                localeutil.StringPhrase("problem.411"),
	http.StatusPreconditionFailed:            localeutil.StringPhrase("problem.412"),
	http.StatusRequestEntityTooLarge:         localeutil.StringPhrase("problem.413"),
	http.StatusRequestURITooLong:             localeutil.StringPhrase("problem.414"),
	http.StatusUnsupportedMediaType:          localeutil.StringPhrase("problem.415"),
	http.StatusRequestedRangeNotSatisfiable:  localeutil.StringPhrase("problem.416"),
	http.StatusExpectationFailed:             localeutil.StringPhrase("problem.417"),
	http.StatusTeapot:                        localeutil.StringPhrase("problem.418"),
	http.StatusMisdirectedRequest:            localeutil.StringPhrase("problem.421"),
	http.StatusUnprocessableEntity:           localeutil.StringPhrase("problem.422"),
	http.StatusLocked:                        localeutil.StringPhrase("problem.423"),
	http.StatusFailedDependency:              localeutil.StringPhrase("problem.424"),
	http.StatusTooEarly:                      localeutil.StringPhrase("problem.425"),
	http.StatusUpgradeRequired:               localeutil.StringPhrase("problem.426"),
	http.StatusPreconditionRequired:          localeutil.StringPhrase("problem.428"),
	http.StatusTooManyRequests:               localeutil.StringPhrase("problem.429"),
	http.StatusRequestHeaderFieldsTooLarge:   localeutil.StringPhrase("problem.431"),
	http.StatusUnavailableForLegalReasons:    localeutil.StringPhrase("problem.451"),
	http.StatusInternalServerError:           localeutil.StringPhrase("problem.500"),
	http.StatusNotImplemented:                localeutil.StringPhrase("problem.501"),
	http.StatusBadGateway:                    localeutil.StringPhrase("problem.502"),
	http.StatusServiceUnavailable:            localeutil.StringPhrase("problem.503"),
	http.StatusGatewayTimeout:                localeutil.StringPhrase("problem.504"),
	http.StatusHTTPVersionNotSupported:       localeutil.StringPhrase("problem.505"),
	http.StatusVariantAlsoNegotiates:         localeutil.StringPhrase("problem.506"),
	http.StatusInsufficientStorage:           localeutil.StringPhrase("problem.507"),
	http.StatusLoopDetected:                  localeutil.StringPhrase("problem.508"),
	http.StatusNotExtended:                   localeutil.StringPhrase("problem.510"),
	http.StatusNetworkAuthenticationRequired: localeutil.StringPhrase("problem.511"),
}

var detailLocales = map[int]localeutil.LocaleStringer{
	http.StatusBadRequest:                    localeutil.StringPhrase("problem.400.detail"),
	http.StatusUnauthorized:                  localeutil.StringPhrase("problem.401.detail"),
	http.StatusPaymentRequired:               localeutil.StringPhrase("problem.402.detail"),
	http.StatusForbidden:                     localeutil.StringPhrase("problem.403.detail"),
	http.StatusNotFound:                      localeutil.StringPhrase("problem.404.detail"),
	http.StatusMethodNotAllowed:              localeutil.StringPhrase("problem.405.detail"),
	http.StatusNotAcceptable:                 localeutil.StringPhrase("problem.406.detail"),
	http.StatusProxyAuthRequired:             localeutil.StringPhrase("problem.407.detail"),
	http.StatusRequestTimeout:                localeutil.StringPhrase("problem.408.detail"),
	http.StatusConflict:                      localeutil.StringPhrase("problem.409.detail"),
	http.StatusGone:                          localeutil.StringPhrase("problem.410.detail"),
	http.StatusLengthRequired:                localeutil.StringPhrase("problem.411.detail"),
	http.StatusPreconditionFailed:            localeutil.StringPhrase("problem.412.detail"),
	http.StatusRequestEntityTooLarge:         localeutil.StringPhrase("problem.413.detail"),
	http.StatusRequestURITooLong:             localeutil.StringPhrase("problem.414.detail"),
	http.StatusUnsupportedMediaType:          localeutil.StringPhrase("problem.415.detail"),
	http.StatusRequestedRangeNotSatisfiable:  localeutil.StringPhrase("problem.416.detail"),
	http.StatusExpectationFailed:             localeutil.StringPhrase("problem.417.detail"),
	http.StatusTeapot:                        localeutil.StringPhrase("problem.418.detail"),
	http.StatusMisdirectedRequest:            localeutil.StringPhrase("problem.421.detail"),
	http.StatusUnprocessableEntity:           localeutil.StringPhrase("problem.422.detail"),
	http.StatusLocked:                        localeutil.StringPhrase("problem.423.detail"),
	http.StatusFailedDependency:              localeutil.StringPhrase("problem.424.detail"),
	http.StatusTooEarly:                      localeutil.StringPhrase("problem.425.detail"),
	http.StatusUpgradeRequired:               localeutil.StringPhrase("problem.426.detail"),
	http.StatusPreconditionRequired:          localeutil.StringPhrase("problem.428.detail"),
	http.StatusTooManyRequests:               localeutil.StringPhrase("problem.429.detail"),
	http.StatusRequestHeaderFieldsTooLarge:   localeutil.StringPhrase("problem.431.detail"),
	http.StatusUnavailableForLegalReasons:    localeutil.StringPhrase("problem.451.detail"),
	http.StatusInternalServerError:           localeutil.StringPhrase("problem.500.detail"),
	http.StatusNotImplemented:                localeutil.StringPhrase("problem.501.detail"),
	http.StatusBadGateway:                    localeutil.StringPhrase("problem.502.detail"),
	http.StatusServiceUnavailable:            localeutil.StringPhrase("problem.503.detail"),
	http.StatusGatewayTimeout:                localeutil.StringPhrase("problem.504.detail"),
	http.StatusHTTPVersionNotSupported:       localeutil.StringPhrase("problem.505.detail"),
	http.StatusVariantAlsoNegotiates:         localeutil.StringPhrase("problem.506.detail"),
	http.StatusInsufficientStorage:           localeutil.StringPhrase("problem.507.detail"),
	http.StatusLoopDetected:                  localeutil.StringPhrase("problem.508.detail"),
	http.StatusNotExtended:                   localeutil.StringPhrase("problem.510.detail"),
	http.StatusNetworkAuthenticationRequired: localeutil.StringPhrase("problem.511.detail"),
}
