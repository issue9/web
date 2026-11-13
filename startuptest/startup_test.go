// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package startuptest

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/issue9/assert/v6"
	"github.com/issue9/cache/drivers/memory"
	"github.com/issue9/config"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/codec"
	"github.com/issue9/web/locales"
)

func newServer(a *assert.Assertion, ts Startup) *web.Server {
	ser := make(config.Serializer, 4)
	ser.Add(xml.Marshal, xml.Unmarshal, ".xml")
	conf, err := config.BuildDir(ser, "@.config")
	a.NotError(err).NotNil(conf)

	cc := memory.New()
	u := unique.NewNumber(100)

	codecInst := web.NewCodec()
	codecInst.AddMediaType(codec.NewJSON(), true, true)

	s := web.NewServer("id", "1.0.0", conf, ts, time.Local, language.SimplifiedChinese, logs.New(logs.NewNopHandler()), u.String, cc, codecInst, header.XRequestID)
	a.NotNil(s)
	s.Services().Add(locales.UniqueIdentityGenerator, u.Serve)

	return s
}

func testRequest(s *web.Server, ts Startup) {
	router := s.Routers().New("main", nil)
	router.Any("/path", func(ctx *web.Context) web.Responser { return web.Created(ctx.Route().Path, "") })

	ts.Get("/path").
		Header(header.Accept, header.JSON).
		Header(header.ContentType, header.JSON).
		Do().
		Status(http.StatusCreated)

	ts.Delete("/path").
		Header(header.Accept, header.JSON).
		Header(header.ContentType, header.JSON).
		Do().
		Status(http.StatusCreated)

	ts.Put("/path").
		Header(header.Accept, header.JSON).
		Header(header.ContentType, header.JSON).
		Do().
		Status(http.StatusCreated)

	ts.Post("/path").
		Header(header.Accept, header.JSON).
		Header(header.ContentType, header.JSON).
		Do().
		Status(http.StatusCreated)

	ts.Patch("/path").
		Header(header.Accept, header.JSON).
		Header(header.ContentType, header.JSON).
		Do().
		Status(http.StatusCreated)
}
