// SPDX-License-Identifier: MIT

package server

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/cache/memory"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"github.com/issue9/mux/v7"
	"golang.org/x/text/language"
)

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
	Cache cache.Cache

	// 日志的输出通道设置
	//
	// 如果此值为空，那么在被初始化 logs.New(nil) 值，表示不会输出到任何通道。
	Logs *logs.Logs

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

	// 每次连接进入时执行的操作
	OnConnection []mux.ConnectionFunc
}

func sanitizeOptions(o *Options) (*Options, error) {
	if o == nil {
		o = &Options{}
	}

	if o.FS == nil {
		dir, err := os.Executable()
		if err != nil {
			return nil, err
		}
		o.FS = os.DirFS(filepath.Dir(dir))
	}

	if o.Location == nil {
		o.Location = time.Local
	}

	if o.Cache == nil {
		o.Cache = memory.New(24 * time.Hour)
	}

	if o.HTTPServer == nil {
		o.HTTPServer = &http.Server{}
	}

	if o.Logs == nil {
		o.Logs = logs.New(nil)
	}

	if o.ProblemBuilder == nil {
		o.ProblemBuilder = RFC7807Builder
	}

	if o.LanguageTag == language.Und {
		tag, err := localeutil.DetectUserLanguageTag()
		if err != nil {
			o.Logs.Error(err) // 输出错误，但是没必要中断程序。
		}
		o.LanguageTag = tag
	}

	return o, nil
}
