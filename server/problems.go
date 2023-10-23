// 此文件由工具产生，请勿手动修改！

package server

import (
	"net/http"

	"github.com/issue9/web"
)

func initProblems(p *problems) {
	p.Add(http.StatusBadRequest, web.LocaleProblem{ID: web.ProblemBadRequest, Title: web.StringPhrase("problem.400"), Detail: web.StringPhrase("problem.400.detail")})
	p.Add(http.StatusUnauthorized, web.LocaleProblem{ID: web.ProblemUnauthorized, Title: web.StringPhrase("problem.401"), Detail: web.StringPhrase("problem.401.detail")})
	p.Add(http.StatusPaymentRequired, web.LocaleProblem{ID: web.ProblemPaymentRequired, Title: web.StringPhrase("problem.402"), Detail: web.StringPhrase("problem.402.detail")})
	p.Add(http.StatusForbidden, web.LocaleProblem{ID: web.ProblemForbidden, Title: web.StringPhrase("problem.403"), Detail: web.StringPhrase("problem.403.detail")})
	p.Add(http.StatusNotFound, web.LocaleProblem{ID: web.ProblemNotFound, Title: web.StringPhrase("problem.404"), Detail: web.StringPhrase("problem.404.detail")})
	p.Add(http.StatusMethodNotAllowed, web.LocaleProblem{ID: web.ProblemMethodNotAllowed, Title: web.StringPhrase("problem.405"), Detail: web.StringPhrase("problem.405.detail")})
	p.Add(http.StatusNotAcceptable, web.LocaleProblem{ID: web.ProblemNotAcceptable, Title: web.StringPhrase("problem.406"), Detail: web.StringPhrase("problem.406.detail")})
	p.Add(http.StatusProxyAuthRequired, web.LocaleProblem{ID: web.ProblemProxyAuthRequired, Title: web.StringPhrase("problem.407"), Detail: web.StringPhrase("problem.407.detail")})
	p.Add(http.StatusRequestTimeout, web.LocaleProblem{ID: web.ProblemRequestTimeout, Title: web.StringPhrase("problem.408"), Detail: web.StringPhrase("problem.408.detail")})
	p.Add(http.StatusConflict, web.LocaleProblem{ID: web.ProblemConflict, Title: web.StringPhrase("problem.409"), Detail: web.StringPhrase("problem.409.detail")})
	p.Add(http.StatusGone, web.LocaleProblem{ID: web.ProblemGone, Title: web.StringPhrase("problem.410"), Detail: web.StringPhrase("problem.410.detail")})
	p.Add(http.StatusLengthRequired, web.LocaleProblem{ID: web.ProblemLengthRequired, Title: web.StringPhrase("problem.411"), Detail: web.StringPhrase("problem.411.detail")})
	p.Add(http.StatusPreconditionFailed, web.LocaleProblem{ID: web.ProblemPreconditionFailed, Title: web.StringPhrase("problem.412"), Detail: web.StringPhrase("problem.412.detail")})
	p.Add(http.StatusRequestEntityTooLarge, web.LocaleProblem{ID: web.ProblemRequestEntityTooLarge, Title: web.StringPhrase("problem.413"), Detail: web.StringPhrase("problem.413.detail")})
	p.Add(http.StatusRequestURITooLong, web.LocaleProblem{ID: web.ProblemRequestURITooLong, Title: web.StringPhrase("problem.414"), Detail: web.StringPhrase("problem.414.detail")})
	p.Add(http.StatusUnsupportedMediaType, web.LocaleProblem{ID: web.ProblemUnsupportedMediaType, Title: web.StringPhrase("problem.415"), Detail: web.StringPhrase("problem.415.detail")})
	p.Add(http.StatusRequestedRangeNotSatisfiable, web.LocaleProblem{ID: web.ProblemRequestedRangeNotSatisfiable, Title: web.StringPhrase("problem.416"), Detail: web.StringPhrase("problem.416.detail")})
	p.Add(http.StatusExpectationFailed, web.LocaleProblem{ID: web.ProblemExpectationFailed, Title: web.StringPhrase("problem.417"), Detail: web.StringPhrase("problem.417.detail")})
	p.Add(http.StatusTeapot, web.LocaleProblem{ID: web.ProblemTeapot, Title: web.StringPhrase("problem.418"), Detail: web.StringPhrase("problem.418.detail")})
	p.Add(http.StatusMisdirectedRequest, web.LocaleProblem{ID: web.ProblemMisdirectedRequest, Title: web.StringPhrase("problem.421"), Detail: web.StringPhrase("problem.421.detail")})
	p.Add(http.StatusUnprocessableEntity, web.LocaleProblem{ID: web.ProblemUnprocessableEntity, Title: web.StringPhrase("problem.422"), Detail: web.StringPhrase("problem.422.detail")})
	p.Add(http.StatusLocked, web.LocaleProblem{ID: web.ProblemLocked, Title: web.StringPhrase("problem.423"), Detail: web.StringPhrase("problem.423.detail")})
	p.Add(http.StatusFailedDependency, web.LocaleProblem{ID: web.ProblemFailedDependency, Title: web.StringPhrase("problem.424"), Detail: web.StringPhrase("problem.424.detail")})
	p.Add(http.StatusTooEarly, web.LocaleProblem{ID: web.ProblemTooEarly, Title: web.StringPhrase("problem.425"), Detail: web.StringPhrase("problem.425.detail")})
	p.Add(http.StatusUpgradeRequired, web.LocaleProblem{ID: web.ProblemUpgradeRequired, Title: web.StringPhrase("problem.426"), Detail: web.StringPhrase("problem.426.detail")})
	p.Add(http.StatusPreconditionRequired, web.LocaleProblem{ID: web.ProblemPreconditionRequired, Title: web.StringPhrase("problem.428"), Detail: web.StringPhrase("problem.428.detail")})
	p.Add(http.StatusTooManyRequests, web.LocaleProblem{ID: web.ProblemTooManyRequests, Title: web.StringPhrase("problem.429"), Detail: web.StringPhrase("problem.429.detail")})
	p.Add(http.StatusRequestHeaderFieldsTooLarge, web.LocaleProblem{ID: web.ProblemRequestHeaderFieldsTooLarge, Title: web.StringPhrase("problem.431"), Detail: web.StringPhrase("problem.431.detail")})
	p.Add(http.StatusUnavailableForLegalReasons, web.LocaleProblem{ID: web.ProblemUnavailableForLegalReasons, Title: web.StringPhrase("problem.451"), Detail: web.StringPhrase("problem.451.detail")})
	p.Add(http.StatusInternalServerError, web.LocaleProblem{ID: web.ProblemInternalServerError, Title: web.StringPhrase("problem.500"), Detail: web.StringPhrase("problem.500.detail")})
	p.Add(http.StatusNotImplemented, web.LocaleProblem{ID: web.ProblemNotImplemented, Title: web.StringPhrase("problem.501"), Detail: web.StringPhrase("problem.501.detail")})
	p.Add(http.StatusBadGateway, web.LocaleProblem{ID: web.ProblemBadGateway, Title: web.StringPhrase("problem.502"), Detail: web.StringPhrase("problem.502.detail")})
	p.Add(http.StatusServiceUnavailable, web.LocaleProblem{ID: web.ProblemServiceUnavailable, Title: web.StringPhrase("problem.503"), Detail: web.StringPhrase("problem.503.detail")})
	p.Add(http.StatusGatewayTimeout, web.LocaleProblem{ID: web.ProblemGatewayTimeout, Title: web.StringPhrase("problem.504"), Detail: web.StringPhrase("problem.504.detail")})
	p.Add(http.StatusHTTPVersionNotSupported, web.LocaleProblem{ID: web.ProblemHTTPVersionNotSupported, Title: web.StringPhrase("problem.505"), Detail: web.StringPhrase("problem.505.detail")})
	p.Add(http.StatusVariantAlsoNegotiates, web.LocaleProblem{ID: web.ProblemVariantAlsoNegotiates, Title: web.StringPhrase("problem.506"), Detail: web.StringPhrase("problem.506.detail")})
	p.Add(http.StatusInsufficientStorage, web.LocaleProblem{ID: web.ProblemInsufficientStorage, Title: web.StringPhrase("problem.507"), Detail: web.StringPhrase("problem.507.detail")})
	p.Add(http.StatusLoopDetected, web.LocaleProblem{ID: web.ProblemLoopDetected, Title: web.StringPhrase("problem.508"), Detail: web.StringPhrase("problem.508.detail")})
	p.Add(http.StatusNotExtended, web.LocaleProblem{ID: web.ProblemNotExtended, Title: web.StringPhrase("problem.510"), Detail: web.StringPhrase("problem.510.detail")})
	p.Add(http.StatusNetworkAuthenticationRequired, web.LocaleProblem{ID: web.ProblemNetworkAuthenticationRequired, Title: web.StringPhrase("problem.511"), Detail: web.StringPhrase("problem.511.detail")})
}
