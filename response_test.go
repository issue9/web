// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/serializer/text"
	"github.com/issue9/web/serializer/text/testobject"
	"github.com/issue9/web/server/servertest"
)

func TestCreated(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	r := s.Router()

	// Location == ""
	r.Get("/created", func(ctx *Context) Responser {
		return Created(&testobject.TextObject{Name: "test", Age: 123}, "")
	})
	// Location == "/test"
	r.Get("/created/location", func(ctx *Context) Responser {
		return Created(&testobject.TextObject{Name: "test", Age: 123}, "/test")
	})

	s.GoServe()

	s.Get("/created").Header("accept", text.Mimetype).Do(nil).
		Status(http.StatusCreated).
		StringBody(`test,123`)

	s.Get("/created/location").Header("accept", text.Mimetype).Do(nil).
		Status(http.StatusCreated).
		StringBody(`test,123`).
		Header("Location", "/test")

	s.Close(0)
}

func TestRedirect(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	r := s.Router()

	r.Get("/not-implement", func(ctx *Context) Responser {
		return ctx.NotImplemented()
	})
	r.Get("/redirect", func(ctx *Context) Responser {
		return Redirect(http.StatusMovedPermanently, "https://example.com")
	})

	s.GoServe()

	s.Get("/not-implement").Do(nil).Status(http.StatusNotImplemented)

	s.Get("/redirect").Do(nil).
		Status(http.StatusOK). // http.Client.Do 会自动重定向并请求
		Header("Location", "")

	s.Close(0)
}
