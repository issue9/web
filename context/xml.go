// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/issue9/logs"
)

const (
	xmlContentType  = "application/xml;charset=utf-8"
	xmlEncodingType = "application/xml"
)

// XMLRender Render 的 XML 编码实现。
//
// NOTE: 会在返回的文件头信息中添加 Content-Type=application/xml;charset=utf-8
// 的信息，若想手动指定该内容，可通过在 headers 中传递同名变量来改变。
func XMLRender(ctx Context, code int, v interface{}, headers map[string]string) {
	accept := ctx.Request().Header.Get("Accept")
	if strings.Index(accept, xmlEncodingType) < 0 && strings.Index(accept, "*/*") < 0 {
		logs.Error("Accept 值不正确：", accept)
		ctx.Response().WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if v == nil {
		xmlRenderHeader(ctx.Response(), code, headers)
		return
	}

	var data []byte
	var err error
	switch val := v.(type) {
	case string:
		data = []byte(val)
	case []byte:
		data = val
	case []rune:
		data = []byte(string(val))
	default:
		if data, err = xml.Marshal(val); err != nil {
			ctx.Response().WriteHeader(http.StatusInternalServerError)
			logs.Error(err)
			return
		}
	}

	xmlRenderHeader(ctx.Response(), code, headers) // NOTE: WriteHeader() 必须在 Write() 之前调用

	// 输出数据
	if _, err = ctx.Response().Write(data); err != nil {
		ctx.Response().WriteHeader(http.StatusInternalServerError)
		logs.Error(err)
		return
	}
}

// 将 headers 当作一个头信息输出，若未指定 Content-Type，
// 则默认添加 application/xml;charset=utf-8 作为其值。
func xmlRenderHeader(w http.ResponseWriter, code int, headers map[string]string) {
	if headers == nil {
		w.Header().Set("Content-Type", xmlContentType)
		w.WriteHeader(code)
		return
	}

	if _, found := headers["Content-Type"]; !found {
		headers["Content-Type"] = xmlContentType
	}

	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(code)
}

// XMLRead Read 的 XML 实现。
func XMLRead(ctx Context, v interface{}) bool {
	if ctx.Request().Method != http.MethodGet {
		ct := ctx.Request().Header.Get("Content-Type")
		if strings.Index(ct, xmlEncodingType) < 0 && strings.Index(ct, "*/*") < 0 {
			ctx.Response().WriteHeader(http.StatusUnsupportedMediaType)
			logs.Error("Content-Type 值不正确：", ct)
			return false
		}
	}

	data, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Response().WriteHeader(http.StatusInternalServerError)
		logs.Error(err)
		return false
	}

	err = xml.Unmarshal(data, v)
	if err != nil {
		ctx.Response().WriteHeader(http.StatusUnprocessableEntity)
		logs.Error(err)
		return false
	}

	return true
}
