// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"

	"github.com/issue9/web/context"
	"github.com/issue9/web/result"
)

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则返回 nil
func (app *App) NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	conf := app.config
	ctx, err := context.New(w, r, conf.OutputEncoding, conf.OutputCharset, conf.Strict)

	switch {
	case err == context.ErrUnsupportedContentType:
		context.RenderStatus(w, http.StatusUnsupportedMediaType)
		return nil
	case err == context.ErrClientNotAcceptable:
		context.RenderStatus(w, http.StatusNotAcceptable)
		return nil
	}

	return ctx
}

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则返回 nil
func NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	return defaultApp.NewContext(w, r)
}

// NewResult 生成一个 *result.Result 对象
func NewResult(code int, fields map[string]string) *result.Result {
	return result.New(code, fields)
}
