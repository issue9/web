// Copyright 2015 by caixw, All rights reserved.
// Use of this source status is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"net/http"
)

// 将v转换成json数据并写入到w中。code为服务端返回的代码。
// 若v的值是string,[]byte,[]rune则直接转换成字符串写入w。
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
	if _, err = w.Write(data); err != nil {
		Error(err)
	}
}
