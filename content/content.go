// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import (
	"errors"
	"net/http"
)

type Renderer interface {
	// Render 向客户端渲染的函数声明。
	//
	// code 为服务端返回的代码；
	// v 为需要输出的变量；
	// headers 用于指定额外的 Header 信息，若传递 nil，则表示没有。
	//
	// NOTE: Render 最终会被 Context.Render 引用，
	// 所以在 Render 的实现者中不能调用 Context.Render 函数
	Render(w http.ResponseWriter, r *http.Request, code int, v interface{}, headers map[string]string)
}

type Reader interface {
	// Read 从客户端读取数据的函数声明。
	//
	// 若成功读取，则返回 true，否则返回 false，并向客户端输出相应的状态码和错误信息。
	//
	// NOTE: Read 最终会被 Context.Read 引用，
	// 所以在 Read 的实现者中不能调用 Context.Read 函数
	Read(w http.ResponseWriter, r *http.Request, v interface{}) bool
}

type Content interface {
	Renderer
	Reader
}

func New(contenttype string, envelopeState int, envelopeKey string, envelopeStatus int) (Content, error) {
	if envelopeState != EnvelopeStateDisable {
		if len(envelopeKey) == 0 {
			return nil, errors.New("无效的 envelopeKey 值")
		}

		if envelopeStatus < 100 || envelopeStatus > 999 {
			return nil, errors.New("无效的 envelopeStatus 值")
		}
	}

	switch contenttype {
	case "json":
		return newJSON(envelopeState, envelopeKey, envelopeStatus), nil
	case "xml":
		return newXML(envelopeState, envelopeKey, envelopeStatus), nil
	default:
		return nil, errors.New("不支持的 contenttype")
	}
}
