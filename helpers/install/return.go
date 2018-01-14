// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package install

// 事件的返回参数类型
type returnType int8

// 事件返回类型
const (
	typeOK returnType = iota
	typeFailed
	typeMessage
)

// Return 事件的返回类型
type Return struct {
	message string
	typ     returnType
}

// ReturnMessage 返回普通信息
func ReturnMessage(msg string) *Return {
	return &Return{
		message: msg,
		typ:     typeMessage,
	}
}

// ReturnError 返回错误信息，内容即 err.Error() 返回的内容
func ReturnError(err error) *Return {
	if err == nil {
		return ReturnOK()
	}

	return &Return{
		message: err.Error(),
		typ:     typeFailed,
	}
}

// ReturnOK 返回正常信息，一般可以直接以 nil 值代替。
func ReturnOK() *Return {
	return &Return{
		typ: typeOK,
	}
}
