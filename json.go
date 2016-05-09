// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/issue9/logs"
)

// RenderJSON 用于将v转换成json数据并写入到w中。code为服务端返回的代码。
// 默认情况下，会在返回的文件头信息中添加Content-Type=application/json;charset=utf-8
// 的信息，若想手动指定该内容，可通过在headers中传递同名变量来改变。
//
// 确保参数w不为nil，否则将触发panic。
// 若v的值是string,[]byte,[]rune则直接转换成字符串写入w。为nil时，
// 不输出任何内容，若需要输出一个空对象，请使用"{}"字符串；
//
// headers用于指定额外的Header信息，若传递nil，则表示没有。
func RenderJSON(w http.ResponseWriter, code int, v interface{}, headers map[string]string) {
	if w == nil {
		panic("web.RenderJSON:参数w不能为空")
	}

	if v == nil {
		renderJSONHeader(w, code, headers)
		return
	}

	var data []byte
	switch val := v.(type) {
	case string:
		data = []byte(val)
	case []byte:
		data = val
	case []rune:
		data = []byte(string(val))
	default:
		var err error
		data, err = json.Marshal(val)
		if err != nil {
			logs.Error("web.RenderJSON:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	renderJSONHeader(w, code, headers)

	// 输出数据
	if _, err := w.Write(data); err != nil {
		logs.Error("web.RenderJSON:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// 将headers当作一个头信息输出，若未指定Content-Type，
// 则默认添加application/json;charset=utf-8作为其值。
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

// ReadJSON 用于将r中的body当作一个json格式的数据读取到v中，
// 若出错，则返回相应的http状态码表示其错误类型。
// 200 表示一切正常。
func ReadJSON(r *http.Request, v interface{}) (status int) {
	if r.Method != "GET" {
		ct := r.Header.Get("Content-Type")
		if strings.Index(ct, "application/json") < 0 && strings.Index(ct, "*/*") < 0 {
			logs.Error("web.ReadJSON:", "request.Content-Type值不正确：", ct)
			return http.StatusUnsupportedMediaType
		}
	}

	accept := r.Header.Get("Accept")
	if strings.Index(accept, "application/json") < 0 && strings.Index(accept, "*/*") < 0 {
		logs.Error("web.ReadJSON:", "request.Accept值不正确：", accept)
		return http.StatusUnsupportedMediaType
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logs.Error("web.ReadJSON:", err)
		return http.StatusInternalServerError
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		logs.Error("web.ReadJSON:", err)
		return 422 // http包中并未定义422错误
	}

	return http.StatusOK
}
