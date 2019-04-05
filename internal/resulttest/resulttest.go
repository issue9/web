// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package resulttest 提供了 app.Result 接口的默认实现，方便测试用。
package resulttest

// New 返回 Result 对象
func New(status, code int, message string) *Result {
	return &Result{
		status:  status,
		Code:    code,
		Message: message,
	}
}

// Result 实现 app.Result 接口内容
type Result struct {
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

// Add app.Result.Add
func (rslt *Result) Add(field, message string) {
	rslt.Detail = append(rslt.Detail, &detail{Field: field, Message: message})
}

// Set app.Result.Set
func (rslt *Result) Set(field, message string) {
	rslt.Detail = append(rslt.Detail, &detail{Field: field, Message: message})
}

// Status app.Result.Status
func (rslt *Result) Status() int {
	return rslt.status
}

// Error app.Result.Error
func (rslt *Result) Error() string {
	return rslt.Message
}
