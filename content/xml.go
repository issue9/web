// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import (
	stdxml "encoding/xml"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/issue9/logs"
)

const (
	xmlContentType  = "application/xml;charset=utf-8"
	xmlEncodingType = "application/xml"
)

// 在将 envelope 解析到 xml 出错时的提示。
// 理论上不会出现此错误，注意保持与 envelope 的导出格式相兼容。
var xmlEnvelopeError = []byte(`<xml><status>500</status><response>服务器出错</response></xml>`)

type xml struct {
	envelopeState int
	conf          *Config
}

func newXML(conf *Config) *xml {
	x := &xml{
		conf: conf,
	}

	switch conf.EnvelopeState {
	case EnvelopeStateMust:
		x.envelopeState = envelopeStateMust
	case EnvelopeStateEnable:
		x.envelopeState = envelopeStateEnable
	case EnvelopeStateDisable:
		x.envelopeState = envelopeStateDisable
	}

	return x
}

func (x *xml) envelope(r *http.Request) bool {
	switch x.envelopeState {
	case envelopeStateDisable:
		return false
	case envelopeStateMust:
		return true
	case envelopeStateEnable:
		return r.FormValue(x.conf.EnvelopeKey) == "true"
	default: // 默认为禁止
		return false
	}
}

func (x *xml) renderEnvelope(w http.ResponseWriter, r *http.Request, code int, resp interface{}) {
	w.WriteHeader(x.conf.EnvelopeStatus)

	accept := r.Header.Get("Accept")
	if strings.Index(accept, xmlEncodingType) < 0 && strings.Index(accept, "*/*") < 0 {
		logs.Error("Accept 值不正确：", accept)
		code = http.StatusNotAcceptable
		resp = nil // 已经出错了，则不输出内容
	}

	e := newEnvelope(code, w.Header(), resp)
	data, err := stdxml.Marshal(e)
	if err != nil {
		logs.Error(err)
		w.Write(xmlEnvelopeError)
		return
	}

	w.Write(data)
}

// Render 将 v 转换成 XML 内容，并向客户端输出。
//
// 若 v 的值是 string,[]byte，[]rune 则直接转换成字符串；为 nil 时，
// 不输出任何内容；若需要输出一个空对象，请使用"{}"字符串；
//
// NOTE: 会在返回的文件头信息中添加 Content-Type=application/xml;charset=utf-8
// 的信息，若想手动指定该内容，可通过在 headers 中传递同名变量来改变。
func (x *xml) Render(w http.ResponseWriter, r *http.Request, code int, v interface{}, headers map[string]string) {
	x.setHeader(w, headers)

	if x.envelope(r) {
		x.renderEnvelope(w, r, code, v)
		return
	}

	accept := r.Header.Get("Accept")
	if strings.Index(accept, xmlEncodingType) < 0 && strings.Index(accept, "*/*") < 0 {
		logs.Error("Accept 值不正确：", accept)
		w.WriteHeader(http.StatusNotAcceptable)
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
		if data, err = stdxml.Marshal(val); err != nil {
			logs.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(code) // NOTE: WriteHeader() 必须在 Write() 之前调用
	if _, err = w.Write(data); err != nil {
		logs.Error(err)
		w.WriteHeader(http.StatusInternalServerError) // BUG(caixw) 会提示重复调用 WriteHeader 的错误
		return
	}
}

// 设置报头内容。若未指定 Content-Type，
// 则默认添加 application/xml;charset=utf-8 作为其值。
func (x *xml) setHeader(w http.ResponseWriter, headers map[string]string) {
	if headers == nil {
		w.Header().Set("Content-Type", xmlContentType)
		return
	}

	if _, found := headers["Content-Type"]; !found {
		headers["Content-Type"] = xmlContentType
	}

	for k, v := range headers {
		w.Header().Set(k, v)
	}
}

// Read 将 r.Body 转换成 XML 对象，并写入到 v 中。
func (x *xml) Read(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if r.Method != http.MethodGet {
		ct := r.Header.Get("Content-Type")
		if strings.Index(ct, xmlEncodingType) < 0 && strings.Index(ct, "*/*") < 0 {
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

	err = stdxml.Unmarshal(data, v)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		logs.Error(err)
		return false
	}

	return true
}
