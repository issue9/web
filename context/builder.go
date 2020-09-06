// SPDX-License-Identifier: MIT

package context

import (
	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/recovery/errorhandler"
)

// Builder 定义了构建 Context 对象的一些通用数据选项
type Builder struct {
	// 对非正常的 HTTP 状态处理会调用此对象输出
	ErrorHandlers *errorhandler.ErrorHandler

	// 可供操作的日志记录
	//
	// 在 Context.Error 等方法中输出错误信息会调用此对象作输出。
	Logs *logs.Logs

	// 在生成 Context 对象之前可对 Context 作的修改操作，
	// 可以为空，表示不作任何操作。
	Interceptor func(ctx *Context)

	// 用于生成 Result 接口对象的函数
	//
	// context 包本身提供了一个默认的实现方式：DefaultResultBuilder
	ResultBuilder BuildResultFunc
	messages      map[int]*resultMessage

	// mimetype
	marshals   []*marshaler
	unmarshals []*unmarshaler
}
