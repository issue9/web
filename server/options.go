// SPDX-License-Identifier: MIT

package server

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v7"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"

	"github.com/issue9/web/cache"
	"github.com/issue9/web/cache/caches"
	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/internal/service"
	"github.com/issue9/web/logs"
)

const RequestIDKey = "X-Request-ID"

// Options [Server] 的初始化参数
//
// 这些参数都有默认值，且无法在 [Server] 初始化之后进行更改。
type Options struct {
	// 项目默认可存取的文件系统
	//
	// 默认情况下为可执行文件所在的目录。
	FS fs.FS

	// 服务器的时区
	//
	// 默认值为 [time.Local]
	Location *time.Location

	// 缓存系统
	//
	// 默认值为内存类型。
	Cache cache.Driver

	// 日志的输出通道设置
	//
	// 如果此值为空，表示不会输出到任何通道。
	Logs *logs.Options

	// 生成 [Problem] 对象的方法
	//
	// 如果为空，那么将采用 [RFC7807Builder] 作为默认值。
	ProblemBuilder BuildProblemFunc

	// 默认的语言标签
	//
	// 在用户请求的报头中没有匹配的语言标签时，会采用此值作为该用户的本地化语言，
	// 同时也用来初始化 [Server.LocalePrinter]。
	//
	// 如果为空，则会尝试读取当前系统的本地化信息。
	LanguageTag language.Tag

	// http.Server 实例的值
	//
	// 如果为空，表示 &http.Server{} 对象。
	HTTPServer *http.Server

	// 生成唯一字符串的方法
	//
	// 供 [Server.UniqueID] 使用。
	//
	// 如果为空，将采用 [unique.NewDate] 作为生成方法，[unique.Date]。
	UniqueGenerator UniqueGenerator

	// 路由选项
	//
	// 将应用 [Server.Routers] 对象之上。
	RoutersOptions []mux.Option

	// 指定获取 x-request-id 内容的报头名
	//
	// 如果为空，则采用 [RequestIDKey] 作为默认值
	RequestIDKey string
}

// UniqueGenerator 唯一 ID 生成器的接口
type UniqueGenerator interface {
	service.Servicer

	// 返回字符串类型的唯一 ID 值
	String() string
}

func sanitizeOptions(o *Options) (*Options, *errs.FieldError) {
	if o == nil {
		o = &Options{}
	}

	if o.FS == nil {
		dir, err := os.Executable()
		if err != nil {
			return nil, errs.NewFieldError("FS", err)
		}
		o.FS = os.DirFS(filepath.Dir(dir))
	}

	if o.Location == nil {
		o.Location = time.Local
	}

	if o.Cache == nil {
		o.Cache = caches.NewMemory(24 * time.Hour)
	}

	if o.HTTPServer == nil {
		o.HTTPServer = &http.Server{}
	}

	if o.Logs == nil {
		o.Logs = &logs.Options{}
	}

	if o.ProblemBuilder == nil {
		o.ProblemBuilder = RFC7807Builder
	}

	if o.UniqueGenerator == nil {
		o.UniqueGenerator = unique.NewDate(1000)
	}

	if o.LanguageTag == language.Und {
		tag, err := localeutil.DetectUserLanguageTag()
		if err != nil {
			return nil, errs.NewFieldError("LanguageTag", err)
		}
		o.LanguageTag = tag
	}

	if o.RequestIDKey == "" {
		o.RequestIDKey = RequestIDKey
	}

	return o, nil
}
