// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"errors"
	"net/http"
	"strings"

	"github.com/issue9/web/context"
	"github.com/issue9/web/encoding"
)

var (
	// ErrExists 表示存在相同名称的项。
	// 多个类似功能的函数，都有可能返回此错误。
	ErrExists = errors.New("存在相同名称的项")
)

// AddMarshal 添加一个新的解码器，只有通过 AddMarshal 添加的解码器，
// 才能被 Context 使用。
func (app *App) AddMarshal(name string, m encoding.Marshal) error {
	_, found := app.marshals[name]
	if found {
		return ErrExists
	}

	app.marshals[name] = m
	return nil
}

// AddUnmarshal 添加一个编码器，只有通过 AddUnmarshal 添加的解码器，
// 才能被 Context 使用。
func (app *App) AddUnmarshal(name string, m encoding.Unmarshal) error {
	_, found := app.unmarshals[name]
	if found {
		return ErrExists
	}

	app.unmarshals[name] = m
	return nil
}

// AddCharset 添加编码方式，只有通过 AddCharset 添加的字符集，
// 才能被 Context 使用。
func (app *App) AddCharset(name string, enc encoding.Charset) error {
	_, found := app.charset[name]
	if found {
		return ErrExists
	}

	app.charset[name] = enc
	return nil
}

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则返回 nil
//
// encodingName 指定出输出时的编码方式，此编码名称必须已经通过 AddMarshal 添加；
// charsetName 指定输出时的字符集，此字符集名称必须已经通过 AddCharset 添加；
// strict 若为 true，则会验证用户的 Accept 报头是否接受 encodingName 编码。
// 输入时的编码与字符集信息从报头 Content-Type 中获取，若未指定字符集，则默认为 utf-8
func (app *App) NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	encName, charsetName := encoding.ParseContentType(r.Header.Get("Content-Type"))

	unmarshal, found := app.unmarshals[encName]
	if !found {
		context.RenderStatus(w, http.StatusUnsupportedMediaType)
		return nil
	}

	inputCharset, found := app.charset[charsetName]
	if !found {
		context.RenderStatus(w, http.StatusUnsupportedMediaType)
		return nil
	}

	if app.config.Strict {
		accept := r.Header.Get("Accept")
		if !strings.Contains(accept, app.outputEncodingName) && !strings.Contains(accept, "*/*") {
			context.RenderStatus(w, http.StatusNotAcceptable)
			return nil
		}
	}

	return &context.Context{
		Response:           w,
		Request:            r,
		OutputEncoding:     app.outputEncoding,
		OutputEncodingName: app.outputEncodingName,
		InputEncoding:      unmarshal,
		InputCharset:       inputCharset,
		OutputCharset:      app.outputCharset,
		OutputCharsetName:  app.outputCharsetName,
	}
}
