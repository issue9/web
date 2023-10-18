// 此文件由工具产生，请勿手动修改！

package server

import (
	"net/http"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/problems"
)

func initProblems(p *problems.Problems) {
	p.Add(web.ProblemBadRequest, http.StatusBadRequest, web.StringPhrase("problem.400"), web.StringPhrase("problem.400.detail"))
	p.Add(web.ProblemUnauthorized, http.StatusUnauthorized, web.StringPhrase("problem.401"), web.StringPhrase("problem.401.detail"))
	p.Add(web.ProblemPaymentRequired, http.StatusPaymentRequired, web.StringPhrase("problem.402"), web.StringPhrase("problem.402.detail"))
	p.Add(web.ProblemForbidden, http.StatusForbidden, web.StringPhrase("problem.403"), web.StringPhrase("problem.403.detail"))
	p.Add(web.ProblemNotFound, http.StatusNotFound, web.StringPhrase("problem.404"), web.StringPhrase("problem.404.detail"))
	p.Add(web.ProblemMethodNotAllowed, http.StatusMethodNotAllowed, web.StringPhrase("problem.405"), web.StringPhrase("problem.405.detail"))
	p.Add(web.ProblemNotAcceptable, http.StatusNotAcceptable, web.StringPhrase("problem.406"), web.StringPhrase("problem.406.detail"))
	p.Add(web.ProblemProxyAuthRequired, http.StatusProxyAuthRequired, web.StringPhrase("problem.407"), web.StringPhrase("problem.407.detail"))
	p.Add(web.ProblemRequestTimeout, http.StatusRequestTimeout, web.StringPhrase("problem.408"), web.StringPhrase("problem.408.detail"))
	p.Add(web.ProblemConflict, http.StatusConflict, web.StringPhrase("problem.409"), web.StringPhrase("problem.409.detail"))
	p.Add(web.ProblemGone, http.StatusGone, web.StringPhrase("problem.410"), web.StringPhrase("problem.410.detail"))
	p.Add(web.ProblemLengthRequired, http.StatusLengthRequired, web.StringPhrase("problem.411"), web.StringPhrase("problem.411.detail"))
	p.Add(web.ProblemPreconditionFailed, http.StatusPreconditionFailed, web.StringPhrase("problem.412"), web.StringPhrase("problem.412.detail"))
	p.Add(web.ProblemRequestEntityTooLarge, http.StatusRequestEntityTooLarge, web.StringPhrase("problem.413"), web.StringPhrase("problem.413.detail"))
	p.Add(web.ProblemRequestURITooLong, http.StatusRequestURITooLong, web.StringPhrase("problem.414"), web.StringPhrase("problem.414.detail"))
	p.Add(web.ProblemUnsupportedMediaType, http.StatusUnsupportedMediaType, web.StringPhrase("problem.415"), web.StringPhrase("problem.415.detail"))
	p.Add(web.ProblemRequestedRangeNotSatisfiable, http.StatusRequestedRangeNotSatisfiable, web.StringPhrase("problem.416"), web.StringPhrase("problem.416.detail"))
	p.Add(web.ProblemExpectationFailed, http.StatusExpectationFailed, web.StringPhrase("problem.417"), web.StringPhrase("problem.417.detail"))
	p.Add(web.ProblemTeapot, http.StatusTeapot, web.StringPhrase("problem.418"), web.StringPhrase("problem.418.detail"))
	p.Add(web.ProblemMisdirectedRequest, http.StatusMisdirectedRequest, web.StringPhrase("problem.421"), web.StringPhrase("problem.421.detail"))
	p.Add(web.ProblemUnprocessableEntity, http.StatusUnprocessableEntity, web.StringPhrase("problem.422"), web.StringPhrase("problem.422.detail"))
	p.Add(web.ProblemLocked, http.StatusLocked, web.StringPhrase("problem.423"), web.StringPhrase("problem.423.detail"))
	p.Add(web.ProblemFailedDependency, http.StatusFailedDependency, web.StringPhrase("problem.424"), web.StringPhrase("problem.424.detail"))
	p.Add(web.ProblemTooEarly, http.StatusTooEarly, web.StringPhrase("problem.425"), web.StringPhrase("problem.425.detail"))
	p.Add(web.ProblemUpgradeRequired, http.StatusUpgradeRequired, web.StringPhrase("problem.426"), web.StringPhrase("problem.426.detail"))
	p.Add(web.ProblemPreconditionRequired, http.StatusPreconditionRequired, web.StringPhrase("problem.428"), web.StringPhrase("problem.428.detail"))
	p.Add(web.ProblemTooManyRequests, http.StatusTooManyRequests, web.StringPhrase("problem.429"), web.StringPhrase("problem.429.detail"))
	p.Add(web.ProblemRequestHeaderFieldsTooLarge, http.StatusRequestHeaderFieldsTooLarge, web.StringPhrase("problem.431"), web.StringPhrase("problem.431.detail"))
	p.Add(web.ProblemUnavailableForLegalReasons, http.StatusUnavailableForLegalReasons, web.StringPhrase("problem.451"), web.StringPhrase("problem.451.detail"))
	p.Add(web.ProblemInternalServerError, http.StatusInternalServerError, web.StringPhrase("problem.500"), web.StringPhrase("problem.500.detail"))
	p.Add(web.ProblemNotImplemented, http.StatusNotImplemented, web.StringPhrase("problem.501"), web.StringPhrase("problem.501.detail"))
	p.Add(web.ProblemBadGateway, http.StatusBadGateway, web.StringPhrase("problem.502"), web.StringPhrase("problem.502.detail"))
	p.Add(web.ProblemServiceUnavailable, http.StatusServiceUnavailable, web.StringPhrase("problem.503"), web.StringPhrase("problem.503.detail"))
	p.Add(web.ProblemGatewayTimeout, http.StatusGatewayTimeout, web.StringPhrase("problem.504"), web.StringPhrase("problem.504.detail"))
	p.Add(web.ProblemHTTPVersionNotSupported, http.StatusHTTPVersionNotSupported, web.StringPhrase("problem.505"), web.StringPhrase("problem.505.detail"))
	p.Add(web.ProblemVariantAlsoNegotiates, http.StatusVariantAlsoNegotiates, web.StringPhrase("problem.506"), web.StringPhrase("problem.506.detail"))
	p.Add(web.ProblemInsufficientStorage, http.StatusInsufficientStorage, web.StringPhrase("problem.507"), web.StringPhrase("problem.507.detail"))
	p.Add(web.ProblemLoopDetected, http.StatusLoopDetected, web.StringPhrase("problem.508"), web.StringPhrase("problem.508.detail"))
	p.Add(web.ProblemNotExtended, http.StatusNotExtended, web.StringPhrase("problem.510"), web.StringPhrase("problem.510.detail"))
	p.Add(web.ProblemNetworkAuthenticationRequired, http.StatusNetworkAuthenticationRequired, web.StringPhrase("problem.511"), web.StringPhrase("problem.511.detail"))
}
