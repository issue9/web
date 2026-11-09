// 此文件由工具产生，请勿手动修改！

package problems

import (
	"net/http"
	"unique"

	"github.com/issue9/localeutil"
)

const AboutBlank = "about:blank"

var (
	Empty = unique.Make("")

	BadRequest                    = unique.Make("400") // RFC 9110, 15.5.1
	Unauthorized                  = unique.Make("401") // RFC 9110, 15.5.2
	PaymentRequired               = unique.Make("402") // RFC 9110, 15.5.3
	Forbidden                     = unique.Make("403") // RFC 9110, 15.5.4
	NotFound                      = unique.Make("404") // RFC 9110, 15.5.5
	MethodNotAllowed              = unique.Make("405") // RFC 9110, 15.5.6
	NotAcceptable                 = unique.Make("406") // RFC 9110, 15.5.7
	ProxyAuthRequired             = unique.Make("407") // RFC 9110, 15.5.8
	RequestTimeout                = unique.Make("408") // RFC 9110, 15.5.9
	Conflict                      = unique.Make("409") // RFC 9110, 15.5.10
	Gone                          = unique.Make("410") // RFC 9110, 15.5.11
	LengthRequired                = unique.Make("411") // RFC 9110, 15.5.12
	PreconditionFailed            = unique.Make("412") // RFC 9110, 15.5.13
	RequestEntityTooLarge         = unique.Make("413") // RFC 9110, 15.5.14
	RequestURITooLong             = unique.Make("414") // RFC 9110, 15.5.15
	UnsupportedMediaType          = unique.Make("415") // RFC 9110, 15.5.16
	RequestedRangeNotSatisfiable  = unique.Make("416") // RFC 9110, 15.5.17
	ExpectationFailed             = unique.Make("417") // RFC 9110, 15.5.18
	Teapot                        = unique.Make("418") // RFC 9110, 15.5.19 (Unused)
	MisdirectedRequest            = unique.Make("421") // RFC 9110, 15.5.20
	UnprocessableEntity           = unique.Make("422") // RFC 9110, 15.5.21
	Locked                        = unique.Make("423") // RFC 4918, 11.3
	FailedDependency              = unique.Make("424") // RFC 4918, 11.4
	TooEarly                      = unique.Make("425") // RFC 8470, 5.2.
	UpgradeRequired               = unique.Make("426") // RFC 9110, 15.5.22
	PreconditionRequired          = unique.Make("428") // RFC 6585, 3
	TooManyRequests               = unique.Make("429") // RFC 6585, 4
	RequestHeaderFieldsTooLarge   = unique.Make("431") // RFC 6585, 5
	UnavailableForLegalReasons    = unique.Make("451") // RFC 7725, 3
	InternalServerError           = unique.Make("500") // RFC 9110, 15.6.1
	NotImplemented                = unique.Make("501") // RFC 9110, 15.6.2
	BadGateway                    = unique.Make("502") // RFC 9110, 15.6.3
	ServiceUnavailable            = unique.Make("503") // RFC 9110, 15.6.4
	GatewayTimeout                = unique.Make("504") // RFC 9110, 15.6.5
	HTTPVersionNotSupported       = unique.Make("505") // RFC 9110, 15.6.6
	VariantAlsoNegotiates         = unique.Make("506") // RFC 2295, 8.1
	InsufficientStorage           = unique.Make("507") // RFC 4918, 11.5
	LoopDetected                  = unique.Make("508") // RFC 5842, 7.2
	NotExtended                   = unique.Make("510") // RFC 2774, 7
	NetworkAuthenticationRequired = unique.Make("511") // RFC 6585, 6

	Date                        = unique.Make("date")                          // RFC 9458, Section 6.5.2
	OhttpKey                    = unique.Make("ohttp-key")                     // RFC 9458, Section 5.3
	DigestUnsupportedAlgorithms = unique.Make("digest-unsupported-algorithms") // RFC-ietf-httpapi-digest-fields-problem-types-06, Section 3.1
	DigestInvalidValues         = unique.Make("digest-invalid-values")         // RFC-ietf-httpapi-digest-fields-problem-types-06, Section 3.2
	DigestMismatchedValues      = unique.Make("digest-mismatched-values")      // RFC-ietf-httpapi-digest-fields-problem-types-06, Section 3.3
)

func (p *Problems) initProblems() {
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

var problemStatuses = map[int]unique.Handle[string]{
	http.StatusBadRequest:                    BadRequest,
	http.StatusUnauthorized:                  Unauthorized,
	http.StatusPaymentRequired:               PaymentRequired,
	http.StatusForbidden:                     Forbidden,
	http.StatusNotFound:                      NotFound,
	http.StatusMethodNotAllowed:              MethodNotAllowed,
	http.StatusNotAcceptable:                 NotAcceptable,
	http.StatusProxyAuthRequired:             ProxyAuthRequired,
	http.StatusRequestTimeout:                RequestTimeout,
	http.StatusConflict:                      Conflict,
	http.StatusGone:                          Gone,
	http.StatusLengthRequired:                LengthRequired,
	http.StatusPreconditionFailed:            PreconditionFailed,
	http.StatusRequestEntityTooLarge:         RequestEntityTooLarge,
	http.StatusRequestURITooLong:             RequestURITooLong,
	http.StatusUnsupportedMediaType:          UnsupportedMediaType,
	http.StatusRequestedRangeNotSatisfiable:  RequestedRangeNotSatisfiable,
	http.StatusExpectationFailed:             ExpectationFailed,
	http.StatusTeapot:                        Teapot,
	http.StatusMisdirectedRequest:            MisdirectedRequest,
	http.StatusUnprocessableEntity:           UnprocessableEntity,
	http.StatusLocked:                        Locked,
	http.StatusFailedDependency:              FailedDependency,
	http.StatusTooEarly:                      TooEarly,
	http.StatusUpgradeRequired:               UpgradeRequired,
	http.StatusPreconditionRequired:          PreconditionRequired,
	http.StatusTooManyRequests:               TooManyRequests,
	http.StatusRequestHeaderFieldsTooLarge:   RequestHeaderFieldsTooLarge,
	http.StatusUnavailableForLegalReasons:    UnavailableForLegalReasons,
	http.StatusInternalServerError:           InternalServerError,
	http.StatusNotImplemented:                NotImplemented,
	http.StatusBadGateway:                    BadGateway,
	http.StatusServiceUnavailable:            ServiceUnavailable,
	http.StatusGatewayTimeout:                GatewayTimeout,
	http.StatusHTTPVersionNotSupported:       HTTPVersionNotSupported,
	http.StatusVariantAlsoNegotiates:         VariantAlsoNegotiates,
	http.StatusInsufficientStorage:           InsufficientStorage,
	http.StatusLoopDetected:                  LoopDetected,
	http.StatusNotExtended:                   NotExtended,
	http.StatusNetworkAuthenticationRequired: NetworkAuthenticationRequired,
}
