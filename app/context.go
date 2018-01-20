// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"
	"strings"

	"github.com/issue9/web/context"
	"github.com/issue9/web/encoding"
)

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则返回 nil
//
// encodingName 指定出输出时的编码方式，此编码名称必须已经通过 AddMarshal 添加；
// charsetName 指定输出时的字符集，此字符集名称必须已经通过 AddCharset 添加；
// strict 若为 true，则会验证用户的 Accept 报头是否接受 encodingName 编码。
// 输入时的编码与字符集信息从报头 Content-Type 中获取，若未指定字符集，则默认为 utf-8
func (app *App) NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	encName, charsetName := encoding.ParseContentType(r.Header.Get("Content-Type"))

	unmarshal, found := app.config.Unmarshals[encName]
	if !found {
		context.RenderStatus(w, http.StatusUnsupportedMediaType)
		return nil
	}

	inputCharset, found := app.config.Charset[charsetName]
	if !found {
		context.RenderStatus(w, http.StatusUnsupportedMediaType)
		return nil
	}

	if app.config.Strict {
		accept := r.Header.Get("Accept")
		if !strings.Contains(accept, app.config.OutputEncoding) && !strings.Contains(accept, "*/*") {
			context.RenderStatus(w, http.StatusNotAcceptable)
			return nil
		}
	}

	return &context.Context{
		Response:           w,
		Request:            r,
		OutputEncoding:     app.config.outputEncoding,
		OutputEncodingName: app.config.OutputEncoding,
		InputEncoding:      unmarshal,
		InputCharset:       inputCharset,
		OutputCharset:      app.config.outputCharset,
		OutputCharsetName:  app.config.OutputCharset,
	}
}
