// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/cache"
	"github.com/issue9/web/cache/caches"
	"github.com/issue9/web/internal/mimetypes"
	"github.com/issue9/web/internal/problems"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/logs"
)

const RequestIDKey = "X-Request-ID"

const DefaultConfigDir = "@.config"

type (
	// Options [Server] 的初始化参数
	//
	// 这些参数都有默认值，且无法在 [Server] 初始化之后进行更改。
	Options struct {
		// 项目的配置项
		//
		// 如果涉及到需要读取配置文件的，可以指定此对象，之后可通过此对象统一处理各类配置文件。
		// 如果为空，则会采用 config.AppDir(DefaultConfigDir) 进行初始化。
		Config *config.Config

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
		logs logs.Logs

		// http.Server 实例的值
		//
		// 如果为空，表示 &http.Server{} 对象。
		HTTPServer *http.Server

		// 生成唯一字符串的方法
		//
		// 供 [Server.UniqueID] 使用。
		//
		// 如果为空，将采用 [unique.NewDate] 作为生成方法，[unique.Date]。
		IDGenerator IDGenerator

		// 路由选项
		//
		// 将应用 [Server.Routers] 对象之上。
		RoutersOptions []RouterOption

		// 指定获取 x-request-id 内容的报头名
		//
		// 如果为空，则采用 [RequestIDKey] 作为默认值
		RequestIDKey string

		// 可用的压缩类型
		//
		// 默认为空。表示不需要该功能。
		Encodings []*Encoding

		// 本地化的相关设置
		//
		// 可以为空，表示根据当前服务器环境检测适当的值。
		Locale *Locale

		// 指定可用的 mimetype
		//
		// 默认为空。
		Mimetypes []*Mimetype
		mimetypes *mimetypes.Mimetypes[MarshalFunc, UnmarshalFunc]

		// ProblemTypePrefix 所有 type 字段的前缀
		//
		// 如果该值为 [ProblemAboutBlank]，将不输出 ID 值；其它值则作为前缀添加。
		ProblemTypePrefix string
		problems          *problems.Problems

		// Init 其它的一些初始化操作
		//
		// 在此可以在用户能实际操作 [Server] 之前对 Server 进行一些操作
		Init []func(*Server)
	}

	Mimetype struct {
		// Mimetype 的值
		Type string

		// 对应的错误状态下的 mimetype 值
		//
		// 可以为空，表示与 Type 相同。
		ProblemType string

		// 编码方法
		Marshal MarshalFunc

		// 解码方法
		Unmarshal UnmarshalFunc
	}

	Encoding struct {
		// 压缩算法的名称
		Name string

		// 压缩算法的构建对象
		Builder NewEncoderFunc

		// 该压缩算法支持的 content-type
		//
		// 如果为空，将被设置为 *
		ContentTypes []string
	}

	// IDGenerator 唯一 ID 生成器的接口
	IDGenerator = func() string

	Locale struct {
		// 默认的语言标签
		//
		// 在用户请求的报头中没有匹配的语言标签时，会采用此值作为该用户的本地化语言，
		// 同时也用来初始化 [Server.LocalePrinter]。
		//
		// 如果为空，则会尝试读取当前系统的本地化信息。
		Language language.Tag

		// 本地化的数据
		//
		// 如果为空，则会被初始化成一个空对象。
		Catalog *catalog.Builder

		printer *message.Printer
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

	if o.Cache == nil {
		o.Cache = caches.NewMemory(24 * time.Hour)
	}

	if o.HTTPServer == nil {
		o.HTTPServer = &http.Server{}
	}

	if o.IDGenerator == nil {
		u := unique.NewDate(1000)
		o.IDGenerator = u.String
		o.Init = append(o.Init, func(s *Server) {
			s.Services().Add(locales.UniqueIdentityGenerator, u)
		})
	}

	if o.Locale == nil {
		o.Locale = &Locale{}
	}
	if err := o.Locale.sanitize(); err != nil {
		return nil, err.AddFieldParent("Locale")
	}

	l, err := logs.New(o.Logs)
	if err != nil {
		return nil, config.NewFieldError("Logs", err)
	}
	o.logs = l

	if o.RequestIDKey == "" {
		o.RequestIDKey = RequestIDKey
	}

	for i, e := range o.Encodings {
		if err := e.sanitize(); err != nil {
			return nil, err.AddFieldParent("Encodings[" + strconv.Itoa(i) + "]")
		}
	}

	// mimetype
	indexes := sliceutil.Dup(o.Mimetypes, func(e1, e2 *Mimetype) bool { return e1.Type == e2.Type })
	if len(indexes) > 0 {
		return nil, config.NewFieldError("Mimetypes["+strconv.Itoa(indexes[0])+"].Type", locales.DuplicateValue)
	}
	o.mimetypes = mimetypes.New[MarshalFunc, UnmarshalFunc](len(o.Mimetypes))
	for _, mt := range o.Mimetypes {
		o.mimetypes.Add(mt.Type, mt.Marshal, mt.Unmarshal, mt.ProblemType)
	}

	o.problems = problems.New(o.ProblemTypePrefix)

	return o, nil
}

func (e *Encoding) sanitize() *config.FieldError {
	if e.Name == "" || e.Name == "identity" || e.Name == "*" {
		return config.NewFieldError("Name", locales.InvalidValue)
	}

	if e.Builder == nil {
		return config.NewFieldError("Builder", locales.CanNotBeEmpty)
	}

	if len(e.ContentTypes) == 0 {
		e.ContentTypes = []string{"*"}
	}

	return nil
}

func (l *Locale) sanitize() *config.FieldError {
	if l.Language == language.Und {
		tag, err := localeutil.DetectUserLanguageTag()
		if err != nil {
			return config.NewFieldError("Language", err)
		}
		l.Language = tag
	}

	if l.Catalog == nil {
		l.Catalog = catalog.NewBuilder(catalog.Fallback(l.Language))
	}

	l.printer = newPrinter(l.Language, l.Catalog)

	return nil
}

func newPrinter(tag language.Tag, cat catalog.Catalog) *message.Printer {
	tag, _, _ = cat.Matcher().Match(tag) // 从 cat 中查找最合适的 tag
	return message.NewPrinter(tag, message.Catalog(cat))
}
