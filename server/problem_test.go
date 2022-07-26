// SPDX-License-Identifier: MIT

package server

import (
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
)

var _ Responser = &Problem{}

func TestProblem_Apply(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	s.Problems().Add("40010", 400, localeutil.Phrase("lang"), localeutil.Phrase("lang"))

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").
		Header("accept", "application/json;charset=utf-8").
		Header("accept-language", language.SimplifiedChinese.String()).
		Request()
	ctx := s.newContext(w, r, nil)
	p := ctx.Problem("40010", nil)
	p.Apply(ctx)
	a.Equal(w.Result().StatusCode, 400)
	body := w.Body.String()
	a.Contains(body, `"title"`).
		Contains(body, `"hans"`).
		Contains(body, "400")
}
