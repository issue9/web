// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

var (
	_ Result        = &ResultData{}
	_ GetResultFunc = getResult
)

func getResult(status, code int, message string) Result {
	return &ResultData{
		status:  status,
		Code:    code,
		Message: message,
	}
}

type ResultData struct {
	XMLName struct{} `json:"-" xml:"result" yaml:"-"`

	// 当前的信息所对应的 HTTP 状态码
	status int

	Message string    `json:"message" xml:"message,attr" yaml:"message"`
	Code    int       `json:"code" xml:"code,attr" yaml:"code"`
	Detail  []*detail `json:"detail,omitempty" xml:"field,omitempty" yaml:"detail,omitempty"`
}

type detail struct {
	Field   string `json:"field" xml:"name,attr" yaml:"field"`
	Message string `json:"message" xml:",chardata" yaml:"message"`
}

func (rslt *ResultData) Add(field, message string) {
	rslt.Detail = append(rslt.Detail, &detail{Field: field, Message: message})
}

func (rslt *ResultData) Set(field, message string) {
	rslt.Detail = append(rslt.Detail, &detail{Field: field, Message: message})
}

func (rslt *ResultData) Status() int {
	return rslt.status
}

func (rslt *ResultData) Error() string {
	return rslt.Message
}
