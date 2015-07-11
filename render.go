// Copyright 2015 by caixw, All rights reserved.
// Use of this source status is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

// 将v转换成json数据并写入到w中。code为服务端返回的代码。
// 若v的值是string,[]byte,[]rune则直接转换成字符串写入w。
// 当v值为nil, "", []byte(""), []rune("")等值时，将输出表示空对象的json字符串："{}"。
// header用于指定额外的Header信息，若传递nil，则表示没有。
func RenderJson(w http.ResponseWriter, code int, v interface{}, header map[string]string) {
	if w == nil {
		panic("参数w不能为空")
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
		if val == nil {
			data = []byte{}
			break
		}

		data, err = json.Marshal(val)
		if err != nil {
			Error(err)
			code = http.StatusInternalServerError
		}
	}

	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	for k, v := range header { // 附加头信息
		w.Header().Add(k, v)
	}

	w.WriteHeader(code)
	if len(data) == 0 {
		data = []byte("{}")
	}
	if _, err = w.Write(data); err != nil {
		Error(err)
	}
}

// 将r中的body当作一个json格式的数据读取到v中，若出错，则直接向w输出出错内容，
// 并返回false，或是在一切正常的情况下返回true
func ReadJson(w http.ResponseWriter, r *http.Request, v interface{}) (ok bool) {
	ct := r.Header.Get("Content-Type")
	if strings.Index(ct, "application/json") < 0 { // 请求格式不正确
		RenderJson(w, http.StatusUnsupportedMediaType, nil, nil)
		return false
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Error(err)
		RenderJson(w, http.StatusInternalServerError, nil, nil)
		return false
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		Error(err)
		RenderJson(w, 422, nil, nil) // http包中并未定义422错误
		return false
	}

	return true
}
