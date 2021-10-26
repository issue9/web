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
	"github.com/issue9/middleware/v5/recovery"
	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/group"
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
	// 默认情况下指向  result.DefaultBuilder。
	ResultBuilder BuildResultFunc

	// 缓存系统
	//
	// 默认值为内存类型。
	Cache cache.Cache

	// 端口号
	//
	// 格式参照 net/http.Server.Addr 字段。可以为空，由 net/http.Server 确定其默认值。
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
	// NOTE: 对 http.Server.Handler 的修改不会启作用，该值始终会指向 Server.groups
	HTTPServer func(*http.Server)
	httpServer *http.Server

	// 在请求崩溃之后的处理方式
	//
	// 这是请求的最后一道防线，如果此函数处理依然 panic，则会造成整个项目退出。
	// 如果为空，则采用 github.com/issue9/middleware/recovery.DefaultRecover。
	Recovery recovery.RecoverFunc

	// 此处列出的类型将不会被压缩
	//
	// 可以带 *，比如 text/* 表示所有 mime-type 为 text/ 开始的类型。
	IgnoreCompressTypes []string

	// 日志的输出通道设置
	//
	// 如果此值为空，那么在被初始化 logs.New(nil) 值，表示不会到任务通道，但是各个函数可用。
	Logs *logs.Logs

	// 指定用于序列化文件的方法
	//
	// 该对象同时被用于加载配置文件和序列化文件。 如果为空，会初始化一个空对象。
	Files *serialization.Files

	// 默认的语言标签
	//
	// 在用户请求的报头中没有匹配的语言标签时，会采用此值作为该用户的本地化语言，
	// 同时也用来初始化 Server.LocalePrinter。
	//
	// 如果为空，则会尝试读取当前系统的本地化信息。
	Tag language.Tag

	locale *serialization.Locale
}

func (o *Options) sanitize() error {
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
		o.ResultBuilder = DefaultBuilder
	}

	if o.Cache == nil {
		o.Cache = memory.New(24 * time.Hour)
	}

	if o.CORS == nil {
		o.CORS = mux.DeniedCORS()
	}
	o.groups = group.New(o.DisableHead, o.CORS)

	o.httpServer = &http.Server{Addr: o.Port}
	if o.HTTPServer != nil {
		o.HTTPServer(o.httpServer)
	}

	if o.Recovery == nil {
		o.Recovery = recovery.DefaultRecover(http.StatusInternalServerError)
	}

	if o.Logs == nil {
		l, err := logs.New(nil)
		if err != nil {
			return err
		}
		o.Logs = l
	}

	if o.Tag == language.Und {
		o.Tag, _ = localeutil.DetectUserLanguageTag()
	}

	if o.Files == nil {
		o.Files = serialization.NewFiles(5)
	}

	b := catalog.NewBuilder(catalog.Fallback(o.Tag))
	o.locale = serialization.NewLocale(b, o.Files)

	return nil
}
