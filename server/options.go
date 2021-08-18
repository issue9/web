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
	"github.com/issue9/logs/v3"
	"github.com/issue9/middleware/v5/recovery"
	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/group"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/content"
	"github.com/issue9/web/serialization"
)

// Options 初始化 Server 的参数
type Options struct {
	// 项目默认可存取的文件系统
	//
	// 默认情况下为可执行文件所在的目录。
	FS fs.FS

	// 服务器的时区
	//
	// 默认值为 time.Local
	Location *time.Location

	// 指定生成 Result 数据的方法
	//
	// 默认情况下指向  result.DefaultBuilder。
	ResultBuilder content.BuildResultFunc

	// 缓存系统
	//
	// 默认值为内存类型。
	Cache cache.Cache

	// 端口号
	//
	// 格式参照 net/http.Server.Addr 字段
	Port string

	// 是否禁止自动生成 HEAD 请求
	DisableHead bool

	// 跨域的相关设置
	//
	// 如果为空，采用 mux.DeniedCORS
	CORS *mux.CORS

	groups *group.Groups

	// 可以对 http.Server 的内容进行修改
	//
	// NOTE: 对 http.Server.Handler 的修改不会启作用，该值始终会指向 Server.mux
	HTTPServer func(*http.Server)
	httpServer *http.Server

	// 在请求崩溃之后的处理方式
	//
	// 这是请求的最后一道防线，如果此函数处理依然 panic，则会造成整个项目退出。
	// 如果为空，则会打印简单的错误堆栈信息。
	Recovery recovery.RecoverFunc

	// 此处列出的类型将不会被压缩
	//
	// 可以带 *，比如 text/* 表示所有 mime-type 为 text/ 开始的类型。
	IgnoreCompressTypes []string

	// 日志的输出通道设置
	//
	// 如果此值为空，那么在被初始化 logs.New(nil) 值，表示不会到任务通道，但是各个函数可用。
	Logs *logs.Logs

	// 指定插件的搜索方式
	//
	// 通过 glob 语法搜索插件，比如：
	//  ~/plugins/*.so
	// 具体可参考：https://golang.org/pkg/path/filepath/#Glob
	// 为空表示没有插件。
	//
	// 当前仅支持部分系统，具体可查看：https://golang.org/pkg/plugin/
	Plugins string

	// 指定用于处理本地化的方法
	//
	// 可以为空。
	Locale *serialization.Locale
}

func (o *Options) sanitize() (*Options, error) {
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

	if o.ResultBuilder == nil {
		o.ResultBuilder = content.DefaultBuilder
	}

	if o.Cache == nil {
		o.Cache = memory.New(24 * time.Hour)
	}

	if o.CORS == nil {
		o.CORS = mux.DeniedCORS()
	}
	o.groups = group.New(o.DisableHead, o.CORS, nil, nil)

	o.httpServer = &http.Server{Addr: o.Port}
	if o.HTTPServer != nil {
		o.HTTPServer(o.httpServer)
	}

	if o.Recovery == nil {
		o.Recovery = recovery.DefaultRecoverFunc(http.StatusInternalServerError)
	}

	if o.Logs == nil {
		l, err := logs.New(nil)
		if err != nil {
			return nil, err
		}
		o.Logs = l
	}

	if o.Locale == nil {
		o.Locale = serialization.NewLocale(catalog.NewBuilder(), serialization.NewFiles(5))
	}

	return o, nil
}
