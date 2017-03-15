// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/issue9/logs"
)

const (
	jsonContentType  = "application/json;charset=utf-8"
	jsonEncodingType = "application/json"
)

// 在将 envelope 解析到 json 出错时的提示。
// 理论上不会出现此错误，注意保持与 envelope 的导出格式相兼容。
var jsonEnvelopeError = []byte(`{"status":500,"response":"服务器错误"}`)

func jsonEnvelopeRender(ctx Context, code int, resp interface{}) {
	w := ctx.Response()
	w.WriteHeader(envelopeStatus) // 固定的报头

	accept := ctx.Request().Header.Get("Accept")
	if strings.Index(accept, jsonEncodingType) < 0 && strings.Index(accept, "*/*") < 0 {
		logs.Error("Accept 值不正确：", accept)
		code = http.StatusUnsupportedMediaType
	}

	e := newEnvelope(code, w.Header(), resp)
	data, err := json.Marshal(e)
	if err != nil {
		logs.Error(err)
		w.Write(jsonEnvelopeError)
		return
	}

	w.Write(data)
}

// JSONRender Render 的 JSON 编码实现。
//
// 若 v 的值是 string,[]byte，[]rune 则直接转换成字符串；为 nil 时，
// 不输出任何内容；若需要输出一个空对象，请使用"{}"字符串；
//
// NOTE: 会在返回的文件头信息中添加 Content-Type=application/json;charset=utf-8
// 的信息，若想手动指定该内容，可通过在 headers 中传递同名变量来改变。
func JSONRender(ctx Context, code int, v interface{}, headers map[string]string) {
	w := ctx.Response()
	jsonSetHeader(w, headers)

	if ctx.Envelope() {
		jsonEnvelopeRender(ctx, code, v)
		return
	}

	accept := ctx.Request().Header.Get("Accept")
	if strings.Index(accept, jsonEncodingType) < 0 && strings.Index(accept, "*/*") < 0 {
		logs.Error("Accept 值不正确：", accept)
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if v == nil {
		w.WriteHeader(code)
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
		if data, err = json.Marshal(val); err != nil {
			logs.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(code) // NOTE: WriteHeader() 必须在 Write() 之前调用
	if _, err = ctx.Response().Write(data); err != nil {
		logs.Error(err)
		ctx.Response().WriteHeader(http.StatusInternalServerError)
		return
	}
}

// 设置报头
func jsonSetHeader(w http.ResponseWriter, headers map[string]string) {
	if headers == nil {
		w.Header().Set("Content-Type", jsonContentType)
		return
	}

	if _, found := headers["Content-Type"]; !found {
		headers["Content-Type"] = jsonContentType
	}

	for k, v := range headers {
		w.Header().Set(k, v)
	}
}

// JSONRead Read 的 JSON 编码实现。
func JSONRead(ctx Context, v interface{}) bool {
	if ctx.Request().Method != http.MethodGet {
		ct := ctx.Request().Header.Get("Content-Type")
		if strings.Index(ct, jsonEncodingType) < 0 && strings.Index(ct, "*/*") < 0 {
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

	err = json.Unmarshal(data, v)
	if err != nil {
		ctx.Response().WriteHeader(http.StatusUnprocessableEntity)
		logs.Error(err)
		return false
	}

	return true
}
