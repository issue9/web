// 此文件由工具产生，请勿手动修改！

package problems

import (
	"net/http"

	"github.com/issue9/localeutil"
)

const (
	AboutBlank = "about:blank"

	BadRequest                    = "400" // RFC 9110, 15.5.1
	Unauthorized                  = "401" // RFC 9110, 15.5.2
	PaymentRequired               = "402" // RFC 9110, 15.5.3
	Forbidden                     = "403" // RFC 9110, 15.5.4
	NotFound                      = "404" // RFC 9110, 15.5.5
	MethodNotAllowed              = "405" // RFC 9110, 15.5.6
	NotAcceptable                 = "406" // RFC 9110, 15.5.7
	ProxyAuthRequired             = "407" // RFC 9110, 15.5.8
	RequestTimeout                = "408" // RFC 9110, 15.5.9
	Conflict                      = "409" // RFC 9110, 15.5.10
	Gone                          = "410" // RFC 9110, 15.5.11
	LengthRequired                = "411" // RFC 9110, 15.5.12
	PreconditionFailed            = "412" // RFC 9110, 15.5.13
	RequestEntityTooLarge         = "413" // RFC 9110, 15.5.14
	RequestURITooLong             = "414" // RFC 9110, 15.5.15
	UnsupportedMediaType          = "415" // RFC 9110, 15.5.16
	RequestedRangeNotSatisfiable  = "416" // RFC 9110, 15.5.17
	ExpectationFailed             = "417" // RFC 9110, 15.5.18
	Teapot                        = "418" // RFC 9110, 15.5.19 (Unused)
	MisdirectedRequest            = "421" // RFC 9110, 15.5.20
	UnprocessableEntity           = "422" // RFC 9110, 15.5.21
	Locked                        = "423" // RFC 4918, 11.3
	FailedDependency              = "424" // RFC 4918, 11.4
	TooEarly                      = "425" // RFC 8470, 5.2.
	UpgradeRequired               = "426" // RFC 9110, 15.5.22
	PreconditionRequired          = "428" // RFC 6585, 3
	TooManyRequests               = "429" // RFC 6585, 4
	RequestHeaderFieldsTooLarge   = "431" // RFC 6585, 5
	UnavailableForLegalReasons    = "451" // RFC 7725, 3
	InternalServerError           = "500" // RFC 9110, 15.6.1
	NotImplemented                = "501" // RFC 9110, 15.6.2
	BadGateway                    = "502" // RFC 9110, 15.6.3
	ServiceUnavailable            = "503" // RFC 9110, 15.6.4
	GatewayTimeout                = "504" // RFC 9110, 15.6.5
	HTTPVersionNotSupported       = "505" // RFC 9110, 15.6.6
	VariantAlsoNegotiates         = "506" // RFC 2295, 8.1
	InsufficientStorage           = "507" // RFC 4918, 11.5
	LoopDetected                  = "508" // RFC 5842, 7.2
	NotExtended                   = "510" // RFC 2774, 7
	NetworkAuthenticationRequired = "511" // RFC 6585, 6

	Date                        = "date"                          // RFC 9458, Section 6.5.2
	OhttpKey                    = "ohttp-key"                     // RFC 9458, Section 5.3
	DigestUnsupportedAlgorithms = "digest-unsupported-algorithms" // RFC-ietf-httpapi-digest-fields-problem-types-06, Section 3.1
	DigestInvalidValues         = "digest-invalid-values"         // RFC-ietf-httpapi-digest-fields-problem-types-06, Section 3.2
	DigestMismatchedValues      = "digest-mismatched-values"      // RFC-ietf-httpapi-digest-fields-problem-types-06, Section 3.3
)

func (p *Problems[T]) initProblems() {
	p.add(&Item{Type: AboutBlank, ID: BadRequest, Title: localeutil.StringPhrase("Bad Request"), Status: http.StatusBadRequest})
	p.add(&Item{Type: AboutBlank, ID: Unauthorized, Title: localeutil.StringPhrase("Unauthorized"), Status: http.StatusUnauthorized})
	p.add(&Item{Type: AboutBlank, ID: PaymentRequired, Title: localeutil.StringPhrase("Payment Required"), Status: http.StatusPaymentRequired})
	p.add(&Item{Type: AboutBlank, ID: Forbidden, Title: localeutil.StringPhrase("Forbidden"), Status: http.StatusForbidden})
	p.add(&Item{Type: AboutBlank, ID: NotFound, Title: localeutil.StringPhrase("Not Found"), Status: http.StatusNotFound})
	p.add(&Item{Type: AboutBlank, ID: MethodNotAllowed, Title: localeutil.StringPhrase("Method Not Allowed"), Status: http.StatusMethodNotAllowed})
	p.add(&Item{Type: AboutBlank, ID: NotAcceptable, Title: localeutil.StringPhrase("Not Acceptable"), Status: http.StatusNotAcceptable})
	p.add(&Item{Type: AboutBlank, ID: ProxyAuthRequired, Title: localeutil.StringPhrase("Proxy Authentication Required"), Status: http.StatusProxyAuthRequired})
	p.add(&Item{Type: AboutBlank, ID: RequestTimeout, Title: localeutil.StringPhrase("Request Timeout"), Status: http.StatusRequestTimeout})
	p.add(&Item{Type: AboutBlank, ID: Conflict, Title: localeutil.StringPhrase("Conflict"), Status: http.StatusConflict})
	p.add(&Item{Type: AboutBlank, ID: Gone, Title: localeutil.StringPhrase("Gone"), Status: http.StatusGone})
	p.add(&Item{Type: AboutBlank, ID: LengthRequired, Title: localeutil.StringPhrase("Length Required"), Status: http.StatusLengthRequired})
	p.add(&Item{Type: AboutBlank, ID: PreconditionFailed, Title: localeutil.StringPhrase("Precondition Failed"), Status: http.StatusPreconditionFailed})
	p.add(&Item{Type: AboutBlank, ID: RequestEntityTooLarge, Title: localeutil.StringPhrase("Request Entity Too Large"), Status: http.StatusRequestEntityTooLarge})
	p.add(&Item{Type: AboutBlank, ID: RequestURITooLong, Title: localeutil.StringPhrase("Request URI Too Long"), Status: http.StatusRequestURITooLong})
	p.add(&Item{Type: AboutBlank, ID: UnsupportedMediaType, Title: localeutil.StringPhrase("Unsupported Media Type"), Status: http.StatusUnsupportedMediaType})
	p.add(&Item{Type: AboutBlank, ID: RequestedRangeNotSatisfiable, Title: localeutil.StringPhrase("Requested Range Not Satisfiable"), Status: http.StatusRequestedRangeNotSatisfiable})
	p.add(&Item{Type: AboutBlank, ID: ExpectationFailed, Title: localeutil.StringPhrase("Expectation Failed"), Status: http.StatusExpectationFailed})
	p.add(&Item{Type: AboutBlank, ID: Teapot, Title: localeutil.StringPhrase("I'm a teapot"), Status: http.StatusTeapot})
	p.add(&Item{Type: AboutBlank, ID: MisdirectedRequest, Title: localeutil.StringPhrase("Misdirected Request"), Status: http.StatusMisdirectedRequest})
	p.add(&Item{Type: AboutBlank, ID: UnprocessableEntity, Title: localeutil.StringPhrase("Unprocessable Entity"), Status: http.StatusUnprocessableEntity})
	p.add(&Item{Type: AboutBlank, ID: Locked, Title: localeutil.StringPhrase("Locked"), Status: http.StatusLocked})
	p.add(&Item{Type: AboutBlank, ID: FailedDependency, Title: localeutil.StringPhrase("Failed Dependency"), Status: http.StatusFailedDependency})
	p.add(&Item{Type: AboutBlank, ID: TooEarly, Title: localeutil.StringPhrase("Too Early"), Status: http.StatusTooEarly})
	p.add(&Item{Type: AboutBlank, ID: UpgradeRequired, Title: localeutil.StringPhrase("Upgrade Required"), Status: http.StatusUpgradeRequired})
	p.add(&Item{Type: AboutBlank, ID: PreconditionRequired, Title: localeutil.StringPhrase("Precondition Required"), Status: http.StatusPreconditionRequired})
	p.add(&Item{Type: AboutBlank, ID: TooManyRequests, Title: localeutil.StringPhrase("Too Many Requests"), Status: http.StatusTooManyRequests})
	p.add(&Item{Type: AboutBlank, ID: RequestHeaderFieldsTooLarge, Title: localeutil.StringPhrase("Request Header Fields Too Large"), Status: http.StatusRequestHeaderFieldsTooLarge})
	p.add(&Item{Type: AboutBlank, ID: UnavailableForLegalReasons, Title: localeutil.StringPhrase("Unavailable For Legal Reasons"), Status: http.StatusUnavailableForLegalReasons})
	p.add(&Item{Type: AboutBlank, ID: InternalServerError, Title: localeutil.StringPhrase("Internal Server Error"), Status: http.StatusInternalServerError})
	p.add(&Item{Type: AboutBlank, ID: NotImplemented, Title: localeutil.StringPhrase("Not Implemented"), Status: http.StatusNotImplemented})
	p.add(&Item{Type: AboutBlank, ID: BadGateway, Title: localeutil.StringPhrase("Bad Gateway"), Status: http.StatusBadGateway})
	p.add(&Item{Type: AboutBlank, ID: ServiceUnavailable, Title: localeutil.StringPhrase("Service Unavailable"), Status: http.StatusServiceUnavailable})
	p.add(&Item{Type: AboutBlank, ID: GatewayTimeout, Title: localeutil.StringPhrase("Gateway Timeout"), Status: http.StatusGatewayTimeout})
	p.add(&Item{Type: AboutBlank, ID: HTTPVersionNotSupported, Title: localeutil.StringPhrase("HTTP Version Not Supported"), Status: http.StatusHTTPVersionNotSupported})
	p.add(&Item{Type: AboutBlank, ID: VariantAlsoNegotiates, Title: localeutil.StringPhrase("Variant Also Negotiates"), Status: http.StatusVariantAlsoNegotiates})
	p.add(&Item{Type: AboutBlank, ID: InsufficientStorage, Title: localeutil.StringPhrase("Insufficient Storage"), Status: http.StatusInsufficientStorage})
	p.add(&Item{Type: AboutBlank, ID: LoopDetected, Title: localeutil.StringPhrase("Loop Detected"), Status: http.StatusLoopDetected})
	p.add(&Item{Type: AboutBlank, ID: NotExtended, Title: localeutil.StringPhrase("Not Extended"), Status: http.StatusNotExtended})
	p.add(&Item{Type: AboutBlank, ID: NetworkAuthenticationRequired, Title: localeutil.StringPhrase("Network Authentication Required"), Status: http.StatusNetworkAuthenticationRequired})

	p.add(&Item{Type: "https://iana.org/assignments/http-problem-types#date", ID: Date, Title: localeutil.StringPhrase("Date Not Acceptable"), Status: 400})
	p.add(&Item{Type: "https://iana.org/assignments/http-problem-types#ohttp-key", ID: OhttpKey, Title: localeutil.StringPhrase("Oblivious HTTP key configuration not acceptable"), Status: 400})
	p.add(&Item{Type: "https://iana.org/assignments/http-problem-types#digest-unsupported-algorithms", ID: DigestUnsupportedAlgorithms, Title: localeutil.StringPhrase("Unsupported Hashing Algorithms"), Status: 400})
	p.add(&Item{Type: "https://iana.org/assignments/http-problem-types#digest-invalid-values", ID: DigestInvalidValues, Title: localeutil.StringPhrase("Invalid Digest Values"), Status: 400})
	p.add(&Item{Type: "https://iana.org/assignments/http-problem-types#digest-mismatched-values", ID: DigestMismatchedValues, Title: localeutil.StringPhrase("Mismatched Digest Values"), Status: 400})
}

var problemStatuses = map[int]struct{}{
	http.StatusBadRequest:                    {},
	http.StatusUnauthorized:                  {},
	http.StatusPaymentRequired:               {},
	http.StatusForbidden:                     {},
	http.StatusNotFound:                      {},
	http.StatusMethodNotAllowed:              {},
	http.StatusNotAcceptable:                 {},
	http.StatusProxyAuthRequired:             {},
	http.StatusRequestTimeout:                {},
	http.StatusConflict:                      {},
	http.StatusGone:                          {},
	http.StatusLengthRequired:                {},
	http.StatusPreconditionFailed:            {},
	http.StatusRequestEntityTooLarge:         {},
	http.StatusRequestURITooLong:             {},
	http.StatusUnsupportedMediaType:          {},
	http.StatusRequestedRangeNotSatisfiable:  {},
	http.StatusExpectationFailed:             {},
	http.StatusTeapot:                        {},
	http.StatusMisdirectedRequest:            {},
	http.StatusUnprocessableEntity:           {},
	http.StatusLocked:                        {},
	http.StatusFailedDependency:              {},
	http.StatusTooEarly:                      {},
	http.StatusUpgradeRequired:               {},
	http.StatusPreconditionRequired:          {},
	http.StatusTooManyRequests:               {},
	http.StatusRequestHeaderFieldsTooLarge:   {},
	http.StatusUnavailableForLegalReasons:    {},
	http.StatusInternalServerError:           {},
	http.StatusNotImplemented:                {},
	http.StatusBadGateway:                    {},
	http.StatusServiceUnavailable:            {},
	http.StatusGatewayTimeout:                {},
	http.StatusHTTPVersionNotSupported:       {},
	http.StatusVariantAlsoNegotiates:         {},
	http.StatusInsufficientStorage:           {},
	http.StatusLoopDetected:                  {},
	http.StatusNotExtended:                   {},
	http.StatusNetworkAuthenticationRequired: {},
}
