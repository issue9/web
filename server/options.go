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
	"github.com/issue9/logs/v3"
	"github.com/issue9/mux/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"

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
	// 默认情况下指向  DefaultResultBuilder。
	ResultBuilder BuildResultFunc

	// 缓存系统
	//
	// 默认值为内存类型。
	Cache cache.Cache

	// 端口号
	//
	// 格式参照 net/http.Server.Addr 字段。
	// 可以为空，表示由 net/http.Server 确定其默认值。
	//
	// NOTE: 该值可能会被 HTTPServer 的操作所覆盖。
	Port string

	// 初始化路由的参数
	//
	// 这些选项会应用在所有的路由上，但是并不是所有选项都起作用，
	// 比如 mux.URLDomain，该值始终是在 NewRouter 中指定。
	// 可以为空。
	RouterOptions []mux.Option

	// 应用于全局的中间件
	//
	// 所有通过 NewRouter 的路由都会应用此中间件。
	Middlewares []MiddlewareFunc

	// 可以对 http.Server 的内容进行修改
	//
	// NOTE: 对 http.Server.Handler 的修改不会启作用，该值始终会指向 Server.groups
	HTTPServer func(*http.Server)
	httpServer *http.Server

	// 日志的输出通道设置
	//
	// 如果此值为空，那么在被初始化 logs.New(nil) 值，表示不会输出到任何通道。
	Logs *logs.Logs

	// 指定用于序列化文件的方法
	//
	// 该对象同时被用于加载配置文件和序列化文件。 如果为空，会初始化一个空对象。
	Files *serialization.Files

	// 忽略的压缩类型
	//
	// 当用户请求的 accept 报头与此列表相匹配时，将不会对此请求的内容进行压缩。
	// 可以有通配符，比如 image/* 表示任意 image/ 开头的内容。
	// 默认为空。
	IgnoreEncodings []string

	// 默认的语言标签
	//
	// 在用户请求的报头中没有匹配的语言标签时，会采用此值作为该用户的本地化语言，
	// 同时也用来初始化 Server.LocalePrinter。
	//
	// 如果为空，则会尝试读取当前系统的本地化信息。
	Tag language.Tag

	// 本地化操作的对象
	//
	// 与 Files 组合构建 serialization.Locale 对象，可以为空。
	Catalog *catalog.Builder

	locale *serialization.Locale
}

func (o *Options) sanitize() (err error) {
	if o.FS == nil {
		dir, err := os.Executable()
		if err != nil {
			return err
		}
		o.FS = os.DirFS(filepath.Dir(dir))
	}

	if o.Location == nil {
		o.Location = time.Local
	}

	if o.ResultBuilder == nil {
		o.ResultBuilder = DefaultResultBuilder
	}

	if o.Cache == nil {
		o.Cache = memory.New(24 * time.Hour)
	}

	o.httpServer = &http.Server{Addr: o.Port}
	if o.HTTPServer != nil {
		o.HTTPServer(o.httpServer)
	}

	if o.Logs == nil {
		l, err := logs.New(nil)
		if err != nil {
			return err
		}
		o.Logs = l
	}

	if o.Tag == language.Und {
		if o.Tag, err = localeutil.DetectUserLanguageTag(); err != nil {
			o.Logs.Error(err) // 输出错误，但是没必要中断程序。
		}
	}

	if o.Files == nil {
		o.Files = serialization.NewFiles(5)
	}

	if o.Catalog == nil {
		o.Catalog = catalog.NewBuilder(catalog.Fallback(o.Tag))
	}

	o.locale = serialization.NewLocale(o.Catalog, o.Files)

	return nil
}
