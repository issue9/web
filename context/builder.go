// SPDX-License-Identifier: MIT

package context

import (
	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/recovery/errorhandler"

	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/result"
)

// Builder 定义了构建 Context 对象的一些通用数据选项
type Builder interface {
	ErrorHandlers() *errorhandler.ErrorHandler
	Mimetypes() *mimetype.Mimetypes
	Logs() *logs.Logs
	Results() *result.Results
	ContextInterceptor(ctx *Context)
}
