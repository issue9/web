// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 提供对 json 数据的输出和读取功能
package json

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	contentType  = "application/json;charset=utf-8"
	encodingType = "application/json"
)

// JSON 是 Encoding 接口的 JSON 操作实现
type JSON struct {
	errlog *log.Logger
}

// New 声明一个新的 JSON 实例
func New(errlog *log.Logger) *JSON {
	return &JSON{
		errlog: errlog,
	}
}

// Render 用于将 v 转换成 json 数据并写入到 w 中。
func (j *JSON) Render(w http.ResponseWriter, code int, v interface{}, headers map[string]string) {
	if v == nil {
		renderHeader(w, code, headers)
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
			j.errlog.Println(err)
			return
		}
	}

	renderHeader(w, code, headers) // NOTE: WriteHeader() 必须在 Write() 之前调用

	// 输出数据
	if _, err = w.Write(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		j.errlog.Println(err)
		return
	}
}

// 将 headers 当作一个头信息输出，若未指定Content-Type，
// 则默认添加 application/json;charset=utf-8 作为其值。
func renderHeader(w http.ResponseWriter, code int, headers map[string]string) {
	if headers == nil {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(code)
		return
	}

	if _, found := headers["Content-Type"]; !found {
		headers["Content-Type"] = contentType
	}

	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(code)
}

// Read 用于将 r 中的 body 当作一个 json 格式的数据读取到 v 中。
// 返回值指定是否出错。若出错，会在函数体中指定出错信息，并将错误代码写入报头。
func (j *JSON) Read(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if r.Method != http.MethodGet {
		ct := r.Header.Get("Content-Type")
		if strings.Index(ct, encodingType) < 0 && strings.Index(ct, "*/*") < 0 {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			j.errlog.Println("Content-Type 值不正确：", ct)
			return false
		}
	}

	accept := r.Header.Get("Accept")
	if strings.Index(accept, encodingType) < 0 && strings.Index(accept, "*/*") < 0 {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		j.errlog.Println("Accept 值不正确：", accept)
		return false
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		j.errlog.Println(err)
		return false
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		j.errlog.Println(err)
		return false
	}

	return true
}
