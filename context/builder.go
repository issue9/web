// SPDX-License-Identifier: MIT

package context

import (
	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/recovery/errorhandler"
)

// Builder 定义了构建 Context 对象的一些通用数据选项
type Builder struct {
	ErrorHandlers      *errorhandler.ErrorHandler
	Logs               *logs.Logs
	ContextInterceptor func(ctx *Context)

	// result
	messages map[int]*resultMessage
	build    BuildResultFunc

	// mimetype
	marshals   []*marshaler
	unmarshals []*unmarshaler
}

// NewBuilder 声明 *Builder 实例
func NewBuilder(b BuildResultFunc) *Builder {
	return &Builder{
		messages: make(map[int]*resultMessage, 100),
		build:    b,

		marshals:   make([]*marshaler, 0, 10),
		unmarshals: make([]*unmarshaler, 0, 10),
	}
}
