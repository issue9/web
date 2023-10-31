// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"time"

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web"
	"github.com/issue9/web/cache"
	"github.com/issue9/web/cache/caches"
	"github.com/issue9/web/codec"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/logs"
)

// RequestIDKey 报头中传递 request id 的报头名称
const RequestIDKey = header.RequestIDKey

// DefaultConfigDir 默认的配置目录地址
const DefaultConfigDir = "@.config"

type (
	// Options [Server] 的初始化参数
	//
	// 这些参数都有默认值，且无法在 [Server] 初始化之后进行更改。
	Options struct {
		// 项目的配置项
		//
		// 如果涉及到需要读取配置文件的，可以指定此对象，之后可通过此对象统一处理各类配置文件。
		// 如果为空，则会采用 config.BuildDir(DefaultConfigDir) 进行初始化。
		Config *config.Config

		// 服务器的时区
		//
		// 默认值为 [time.Local]
		Location *time.Location

		// 缓存系统
		//
		// 默认值为内存类型。
		Cache cache.Driver

		// 日志的相关设置
		//
		// 如果此值为空，表示不会输出任何信息。
		Logs *logs.Options
		logs logs.Logs

		// http.Server 实例的值
		//
		// 可以为零值。
		HTTPServer *http.Server

		// 生成唯一字符串的方法
		//
		// 供 [Server.UniqueID] 使用。
		//
		// 如果为空，将采用 [unique.NewDate] 作为生成方法。
		IDGenerator IDGenerator

		// 路由选项
		RoutersOptions []web.RouterOption

		// 构建 [web.Context] 对象时的一此设置
		Context *Context

		// 可用的压缩类型
		//
		// 默认为空。表示不需要该功能。
		Compressions []*codec.Compression

		// 指定可用的 mimetype
		//
		// 默认为空。
		Mimetypes []*codec.Mimetype

		codec web.Codec // 由 Compressions 和 Mimetypes 形成

		// 默认的语言标签
		//
		// 在用户请求的报头中没有匹配的语言标签时，会采用此值作为该用户的本地化语言，
		// 同时也用来初始化 [Server.LocalePrinter]。
		//
		// 框架中的日志输出时，如果该信息实现了 [LocaleStringer] 接口，
		// 将会转换成此设置项的语言。
		//
		// 如果为空，则会尝试读取当前系统的本地化信息。
		Language language.Tag

		// 本地化的数据
		//
		// 如果为空，则会被初始化成一个空对象。
		Catalog *catalog.Builder

		printer *message.Printer // 由 Language 和 Catalog 形成

		// ProblemTypePrefix 所有 type 字段的前缀
		//
		// 如果该值为 [ProblemAboutBlank]，将不输出 ID 值；其它值则作为前缀添加。
		ProblemTypePrefix string
		problems          *problems

		// Init 其它的一些初始化操作
		//
		// 在此可以在用户能实际操作 [Server] 之前对 Server 进行一些操作。
		Init []func(web.Server)
	}

	// IDGenerator 生成唯一 ID 的函数
	IDGenerator = func() string

	Context struct {
		// 指定获取 x-request-id 内容的报头名
		//
		// 如果为空，则采用 [RequestIDKey] 作为默认值
		RequestIDKey string

		// 生成与 [web.Context.Logs] 的固定字段
		//
		// 具体可参考 [web.NewContextBuilder] 的参数说明；
		Logs func(*web.Context) map[string]any
	}
)

func sanitizeOptions(o *Options) (*Options, *config.FieldError) {
	if o == nil {
		o = &Options{}
	}

	if o.Config == nil {
		cfg, err := config.BuildDir(nil, DefaultConfigDir)
		if err != nil {
			return nil, config.NewFieldError("Config", err)
		}
		o.Config = cfg
	}

	if o.Location == nil {
		o.Location = time.Local
	}

	if o.HTTPServer == nil {
		o.HTTPServer = &http.Server{}
	}

	if o.IDGenerator == nil {
		u := unique.NewDate(1000)
		o.IDGenerator = u.String
		o.Init = append(o.Init, func(s web.Server) {
			s.Services().Add(locales.UniqueIdentityGenerator, u)
		})
	}

	if o.Cache == nil {
		c, job := caches.NewMemory()
		o.Cache = c
		o.Init = append(o.Init, func(s web.Server) { // AddTicker 依赖 IDGenerator
			s.Services().AddTicker(locales.RecycleLocalCache, job, time.Minute, false, false)
		})
	}

	if o.Language == language.Und {
		tag, err := localeutil.DetectUserLanguageTag()
		if err != nil {
			return nil, config.NewFieldError("Language", err)
		}
		o.Language = tag
	}

	if o.Catalog == nil {
		o.Catalog = catalog.NewBuilder(catalog.Fallback(o.Language))
	}

	o.printer = newPrinter(o.Language, o.Catalog)

	l, err := logs.New(o.printer, o.Logs)
	if err != nil {
		return nil, config.NewFieldError("Logs", err)
	}
	o.logs = l

	if o.Context == nil {
		o.Context = &Context{}
	}
	o.Context.sanitize()

	c, fe := codec.New("Mimetypes", "Compressions", o.Mimetypes, o.Compressions)
	if err != nil {
		return nil, fe
	}
	o.codec = c

	o.problems = newProblems(o.ProblemTypePrefix)

	return o, nil
}

func (c *Context) sanitize() {
	if c.RequestIDKey == "" {
		c.RequestIDKey = RequestIDKey
	}

	if c.Logs == nil {
		c.Logs = func(ctx *web.Context) map[string]any {
			return map[string]any{
				c.RequestIDKey: ctx.ID(),
			}
		}
	}
}

func newPrinter(tag language.Tag, cat catalog.Catalog) *message.Printer {
	tag, _, _ = cat.Matcher().Match(tag) // 从 cat 中查找最合适的 tag
	return message.NewPrinter(tag, message.Catalog(cat))
}
