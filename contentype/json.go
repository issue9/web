// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package contentype

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

// jsonType 是 ContentTyper 接口的 jsonType 操作实现
type jsonType struct {
}

func newJSON() *jsonType {
	return &jsonType{}
}

// Render 用于将 v 转换成 json 数据并写入到 w 中。
//
// 若 v 的值是 string,[]byte，[]rune 则直接转换成字符串写入 w。为 nil 时，
// 不输出任何内容，若需要输出一个空对象，请使用"{}"字符串；
//
// NOTE: 会在返回的文件头信息中添加 Content-Type=application/json;charset=utf-8
// 的信息，若想手动指定该内容，可通过在 headers 中传递同名变量来改变。
func (j *jsonType) Render(w http.ResponseWriter, r *http.Request, code int, v interface{}, headers map[string]string) {
	accept := r.Header.Get("Accept")
	if strings.Index(accept, jsonEncodingType) < 0 && strings.Index(accept, "*/*") < 0 {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		logs.Error("Accept 值不正确：", accept)
		return
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
			logs.Error(err)
			return
		}
	}

	renderJSONHeader(w, code, headers) // NOTE: WriteHeader() 必须在 Write() 之前调用

	// 输出数据
	if _, err = w.Write(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logs.Error(err)
		return
	}
}

// 将 headers 当作一个头信息输出，若未指定 Content-Type，
// 则默认添加 application/json;charset=utf-8 作为其值。
func renderJSONHeader(w http.ResponseWriter, code int, headers map[string]string) {
	if headers == nil {
		w.Header().Set("Content-Type", jsonContentType)
		w.WriteHeader(code)
		return
	}

	if _, found := headers["Content-Type"]; !found {
		headers["Content-Type"] = jsonContentType
	}

	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(code)
}

// Read 用于将 r 中的 body 当作一个 json 格式的数据读取到 v 中。
// 返回值指定是否出错。若出错，会在函数体中指定出错信息，并将错误代码写入报头。
func (j *jsonType) Read(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if r.Method != http.MethodGet {
		ct := r.Header.Get("Content-Type")
		if strings.Index(ct, jsonEncodingType) < 0 && strings.Index(ct, "*/*") < 0 {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			logs.Error("Content-Type 值不正确：", ct)
			return false
		}
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logs.Error(err)
		return false
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		logs.Error(err)
		return false
	}

	return true
}
