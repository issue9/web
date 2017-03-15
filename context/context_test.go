// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import "net/http"

var _ Context = &defaultContext{}

type defaultContext struct {
	w http.ResponseWriter
	r *http.Request

	envelope bool
}

func (ctx *defaultContext) Render(code int, v interface{}, headers map[string]string) {
}

func (ctx *defaultContext) Read(v interface{}) bool {
	return true
}

func (ctx *defaultContext) Response() http.ResponseWriter {
	return ctx.w
}

func (ctx *defaultContext) Request() *http.Request {
	return ctx.r
}

func (ctx *defaultContext) Envelope() bool {
	return ctx.envelope
}

func newDefaultContext(w http.ResponseWriter, r *http.Request) *defaultContext {
	return &defaultContext{
		w: w,
		r: r,
	}
}
