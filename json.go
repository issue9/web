// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

// RenderJSON 用于将 v 转换成 json 数据并写入到 w 中。
//
// code 为服务端返回的代码。
//
// 确保参数 w 不为 nil，否则将触发 panic。
//
// 若 v 的值是 string,[]byte，[]rune 则直接转换成字符串写入 w。为 nil 时，
// 不输出任何内容，若需要输出一个空对象，请使用"{}"字符串；
//
// headers 用于指定额外的 Header 信息，若传递 nil，则表示没有。
// NOTE: 会在返回的文件头信息中添加 Content-Type=application/json;charset=utf-8
// 的信息，若想手动指定该内容，可通过在 headers 中传递同名变量来改变。
func RenderJSON(w http.ResponseWriter, r *http.Request, code int, v interface{}, headers map[string]string) {
	if w == nil {
		panic("web.RenderJSON:参数w不能为空")
	}

	if v == nil {
		renderJSONHeader(w, code, headers)
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
			w.WriteHeader(http.StatusInternalServerError)
			Error(r, "web.RenderJSON:", err)
			return
		}
	}

	renderJSONHeader(w, code, headers)

	// 输出数据
	if _, err = w.Write(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Error(r, "web.RenderJSON:", err)
		return
	}
}

// 将 headers 当作一个头信息输出，若未指定Content-Type，
// 则默认添加 application/json;charset=utf-8 作为其值。
func renderJSONHeader(w http.ResponseWriter, code int, headers map[string]string) {
	if headers == nil {
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		w.WriteHeader(code)
		return
	}

	if _, found := headers["Content-Type"]; !found {
		headers["Content-Type"] = "application/json;charset=utf-8"
	}

	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(code)
}

// ReadJSON 用于将 r 中的 body 当作一个 json 格式的数据读取到 v 中。
// 返回值指定是否出错。若出错，会在函数体中指定出错信息，并将错误代码写入报头。
func ReadJSON(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if r.Method != "GET" {
		ct := r.Header.Get("Content-Type")
		if strings.Index(ct, "application/json") < 0 && strings.Index(ct, "*/*") < 0 {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			Error(r, "web.ReadJSON:", "request.Content-Type值不正确：", ct)
			return false
		}
	}

	accept := r.Header.Get("Accept")
	if strings.Index(accept, "application/json") < 0 && strings.Index(accept, "*/*") < 0 {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		Error(r, "web.ReadJSON:", "request.Accept值不正确：", accept)
		return false
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Error(r, "web.ReadJSON:", err)
		return false
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		w.WriteHeader(422) // http 包中并未定义 422 错误
		Error(r, "web.ReadJSON:", err)
		return false
	}

	return true
}
