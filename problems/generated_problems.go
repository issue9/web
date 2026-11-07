// 此文件由工具产生，请勿手动修改！

package problems

import (
	"net/http"

	"github.com/issue9/localeutil"
)

const (
	AboutBlank = "about:blank"

	BadRequest                    = "400"
	Unauthorized                  = "401"
	PaymentRequired               = "402"
	Forbidden                     = "403"
	NotFound                      = "404"
	MethodNotAllowed              = "405"
	NotAcceptable                 = "406"
	ProxyAuthRequired             = "407"
	RequestTimeout                = "408"
	Conflict                      = "409"
	Gone                          = "410"
	LengthRequired                = "411"
	PreconditionFailed            = "412"
	RequestEntityTooLarge         = "413"
	RequestURITooLong             = "414"
	UnsupportedMediaType          = "415"
	RequestedRangeNotSatisfiable  = "416"
	ExpectationFailed             = "417"
	Teapot                        = "418"
	MisdirectedRequest            = "421"
	UnprocessableEntity           = "422"
	Locked                        = "423"
	FailedDependency              = "424"
	TooEarly                      = "425"
	UpgradeRequired               = "426"
	PreconditionRequired          = "428"
	TooManyRequests               = "429"
	RequestHeaderFieldsTooLarge   = "431"
	UnavailableForLegalReasons    = "451"
	InternalServerError           = "500"
	NotImplemented                = "501"
	BadGateway                    = "502"
	ServiceUnavailable            = "503"
	GatewayTimeout                = "504"
	HTTPVersionNotSupported       = "505"
	VariantAlsoNegotiates         = "506"
	InsufficientStorage           = "507"
	LoopDetected                  = "508"
	NotExtended                   = "510"
	NetworkAuthenticationRequired = "511"
)

func (p *Problems[T]) initProblems() {
	p.add(&Item{Type: "about:blank", ID: BadRequest, Title: localeutil.StringPhrase("problem.400"), Status: http.StatusBadRequest})
	p.add(&Item{Type: "about:blank", ID: Unauthorized, Title: localeutil.StringPhrase("problem.401"), Status: http.StatusUnauthorized})
	p.add(&Item{Type: "about:blank", ID: PaymentRequired, Title: localeutil.StringPhrase("problem.402"), Status: http.StatusPaymentRequired})
	p.add(&Item{Type: "about:blank", ID: Forbidden, Title: localeutil.StringPhrase("problem.403"), Status: http.StatusForbidden})
	p.add(&Item{Type: "about:blank", ID: NotFound, Title: localeutil.StringPhrase("problem.404"), Status: http.StatusNotFound})
	p.add(&Item{Type: "about:blank", ID: MethodNotAllowed, Title: localeutil.StringPhrase("problem.405"), Status: http.StatusMethodNotAllowed})
	p.add(&Item{Type: "about:blank", ID: NotAcceptable, Title: localeutil.StringPhrase("problem.406"), Status: http.StatusNotAcceptable})
	p.add(&Item{Type: "about:blank", ID: ProxyAuthRequired, Title: localeutil.StringPhrase("problem.407"), Status: http.StatusProxyAuthRequired})
	p.add(&Item{Type: "about:blank", ID: RequestTimeout, Title: localeutil.StringPhrase("problem.408"), Status: http.StatusRequestTimeout})
	p.add(&Item{Type: "about:blank", ID: Conflict, Title: localeutil.StringPhrase("problem.409"), Status: http.StatusConflict})
	p.add(&Item{Type: "about:blank", ID: Gone, Title: localeutil.StringPhrase("problem.410"), Status: http.StatusGone})
	p.add(&Item{Type: "about:blank", ID: LengthRequired, Title: localeutil.StringPhrase("problem.411"), Status: http.StatusLengthRequired})
	p.add(&Item{Type: "about:blank", ID: PreconditionFailed, Title: localeutil.StringPhrase("problem.412"), Status: http.StatusPreconditionFailed})
	p.add(&Item{Type: "about:blank", ID: RequestEntityTooLarge, Title: localeutil.StringPhrase("problem.413"), Status: http.StatusRequestEntityTooLarge})
	p.add(&Item{Type: "about:blank", ID: RequestURITooLong, Title: localeutil.StringPhrase("problem.414"), Status: http.StatusRequestURITooLong})
	p.add(&Item{Type: "about:blank", ID: UnsupportedMediaType, Title: localeutil.StringPhrase("problem.415"), Status: http.StatusUnsupportedMediaType})
	p.add(&Item{Type: "about:blank", ID: RequestedRangeNotSatisfiable, Title: localeutil.StringPhrase("problem.416"), Status: http.StatusRequestedRangeNotSatisfiable})
	p.add(&Item{Type: "about:blank", ID: ExpectationFailed, Title: localeutil.StringPhrase("problem.417"), Status: http.StatusExpectationFailed})
	p.add(&Item{Type: "about:blank", ID: Teapot, Title: localeutil.StringPhrase("problem.418"), Status: http.StatusTeapot})
	p.add(&Item{Type: "about:blank", ID: MisdirectedRequest, Title: localeutil.StringPhrase("problem.421"), Status: http.StatusMisdirectedRequest})
	p.add(&Item{Type: "about:blank", ID: UnprocessableEntity, Title: localeutil.StringPhrase("problem.422"), Status: http.StatusUnprocessableEntity})
	p.add(&Item{Type: "about:blank", ID: Locked, Title: localeutil.StringPhrase("problem.423"), Status: http.StatusLocked})
	p.add(&Item{Type: "about:blank", ID: FailedDependency, Title: localeutil.StringPhrase("problem.424"), Status: http.StatusFailedDependency})
	p.add(&Item{Type: "about:blank", ID: TooEarly, Title: localeutil.StringPhrase("problem.425"), Status: http.StatusTooEarly})
	p.add(&Item{Type: "about:blank", ID: UpgradeRequired, Title: localeutil.StringPhrase("problem.426"), Status: http.StatusUpgradeRequired})
	p.add(&Item{Type: "about:blank", ID: PreconditionRequired, Title: localeutil.StringPhrase("problem.428"), Status: http.StatusPreconditionRequired})
	p.add(&Item{Type: "about:blank", ID: TooManyRequests, Title: localeutil.StringPhrase("problem.429"), Status: http.StatusTooManyRequests})
	p.add(&Item{Type: "about:blank", ID: RequestHeaderFieldsTooLarge, Title: localeutil.StringPhrase("problem.431"), Status: http.StatusRequestHeaderFieldsTooLarge})
	p.add(&Item{Type: "about:blank", ID: UnavailableForLegalReasons, Title: localeutil.StringPhrase("problem.451"), Status: http.StatusUnavailableForLegalReasons})
	p.add(&Item{Type: "about:blank", ID: InternalServerError, Title: localeutil.StringPhrase("problem.500"), Status: http.StatusInternalServerError})
	p.add(&Item{Type: "about:blank", ID: NotImplemented, Title: localeutil.StringPhrase("problem.501"), Status: http.StatusNotImplemented})
	p.add(&Item{Type: "about:blank", ID: BadGateway, Title: localeutil.StringPhrase("problem.502"), Status: http.StatusBadGateway})
	p.add(&Item{Type: "about:blank", ID: ServiceUnavailable, Title: localeutil.StringPhrase("problem.503"), Status: http.StatusServiceUnavailable})
	p.add(&Item{Type: "about:blank", ID: GatewayTimeout, Title: localeutil.StringPhrase("problem.504"), Status: http.StatusGatewayTimeout})
	p.add(&Item{Type: "about:blank", ID: HTTPVersionNotSupported, Title: localeutil.StringPhrase("problem.505"), Status: http.StatusHTTPVersionNotSupported})
	p.add(&Item{Type: "about:blank", ID: VariantAlsoNegotiates, Title: localeutil.StringPhrase("problem.506"), Status: http.StatusVariantAlsoNegotiates})
	p.add(&Item{Type: "about:blank", ID: InsufficientStorage, Title: localeutil.StringPhrase("problem.507"), Status: http.StatusInsufficientStorage})
	p.add(&Item{Type: "about:blank", ID: LoopDetected, Title: localeutil.StringPhrase("problem.508"), Status: http.StatusLoopDetected})
	p.add(&Item{Type: "about:blank", ID: NotExtended, Title: localeutil.StringPhrase("problem.510"), Status: http.StatusNotExtended})
	p.add(&Item{Type: "about:blank", ID: NetworkAuthenticationRequired, Title: localeutil.StringPhrase("problem.511"), Status: http.StatusNetworkAuthenticationRequired})
}
