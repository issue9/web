// SPDX-License-Identifier: MIT

package web

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"net/http"
	"strconv"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/cache"
	"github.com/issue9/web/cache/caches"
	"github.com/issue9/web/internal/compress"
	"github.com/issue9/web/internal/mimetypes"
	"github.com/issue9/web/internal/problems"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/logs"
)

// RequestIDKey 报头中传递 request id 的报头名称
const RequestIDKey = "X-Request-ID"

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
		logs Logs

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
		RoutersOptions []RouterOption

		// 指定获取 x-request-id 内容的报头名
		//
		// 如果为空，则采用 [RequestIDKey] 作为默认值
		RequestIDKey string

		// 可用的压缩类型
		//
		// 默认为空。表示不需要该功能。
		Compresses []*Compress
		compresses *compress.Compresses

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

		// 指定可用的 mimetype
		//
		// 默认为空。
		Mimetypes []*Mimetype
		mimetypes *mtsType

		// ProblemTypePrefix 所有 type 字段的前缀
		//
		// 如果该值为 [ProblemAboutBlank]，将不输出 ID 值；其它值则作为前缀添加。
		ProblemTypePrefix string
		problems          *problems.Problems

		// Init 其它的一些初始化操作
		//
		// 在此可以在用户能实际操作 [Server] 之前对 Server 进行一些操作。
		Init []func(*Server)
	}

	Mimetype struct {
		// Mimetype 的值
		Type string

		// 对应的错误状态下的 mimetype 值
		//
		// 可以为空，表示与 Type 相同。
		ProblemType string

		// 生成编码方法
		MarshalBuilder BuildMarshalFunc

		// 解码方法
		Unmarshal UnmarshalFunc
	}

	// Compress 压缩算法的配置
	Compress struct {
		// Name 压缩方法的名称
		//
		// 可以重名，比如 gzip，可以配置参数不同的对象。
		Name string

		// Compressor 压缩对象
		Compressor Compressor

		// Types 该压缩对象允许使用的为 content-type 类型
		//
		// 如果是 * 表示适用所有类型。
		Types []string
	}

	// Compressor 压缩算法的接口
	Compressor = compress.Compressor

	// IDGenerator 生成唯一 ID 的函数
	IDGenerator = func() string
)

func sanitizeOptions(o *Options) (*Options, *FieldError) {
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
		o.Init = append(o.Init, func(s *Server) {
			s.Services().Add(locales.UniqueIdentityGenerator, u)
		})
	}

	if o.Cache == nil {
		c, job := caches.NewMemory()
		o.Cache = c
		o.Init = append(o.Init, func(s *Server) { // AddTicker 依赖 IDGenerator
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

	l, err := logs.New(o.Logs)
	if err != nil {
		return nil, config.NewFieldError("Logs", err)
	}
	o.logs = l

	if o.RequestIDKey == "" {
		o.RequestIDKey = RequestIDKey
	}

	o.compresses = compress.NewCompresses(len(o.Compresses), false)
	for i, e := range o.Compresses {
		if err := e.sanitize(); err != nil {
			return nil, err.AddFieldParent("Encodings[" + strconv.Itoa(i) + "]")
		}
		o.compresses.Add(e.Name, e.Compressor, e.Types...)
	}

	// mimetype
	indexes := sliceutil.Dup(o.Mimetypes, func(e1, e2 *Mimetype) bool { return e1.Type == e2.Type })
	if len(indexes) > 0 {
		return nil, config.NewFieldError("Mimetypes["+strconv.Itoa(indexes[0])+"].Type", locales.DuplicateValue)
	}
	o.mimetypes = mimetypes.New[BuildMarshalFunc, UnmarshalFunc](len(o.Mimetypes))
	for _, mt := range o.Mimetypes {
		o.mimetypes.Add(mt.Type, mt.MarshalBuilder, mt.Unmarshal, mt.ProblemType)
	}

	o.problems = problems.New(o.ProblemTypePrefix)

	return o, nil
}

func (e *Compress) sanitize() *FieldError {
	if e.Name == "" || e.Name == compress.Identity || e.Name == "*" {
		return config.NewFieldError("Name", locales.InvalidValue)
	}

	if e.Compressor == nil {
		return config.NewFieldError("Compress", locales.CanNotBeEmpty)
	}

	if len(e.Types) == 0 {
		e.Types = []string{"*"}
	}

	return nil
}

func newPrinter(tag language.Tag, cat catalog.Catalog) *message.Printer {
	tag, _, _ = cat.Matcher().Match(tag) // 从 cat 中查找最合适的 tag
	return message.NewPrinter(tag, message.Catalog(cat))
}

// NewZstdCompress 声明基于 [zstd] 的压缩算法
//
// NOTE: 请注意[浏览器支持情况]
//
// [浏览器支持情况]: https://caniuse.com/zstd
// [zstd]: https://www.rfc-editor.org/rfc/rfc8878.html
func NewZstdCompress() Compressor { return compress.NewZstdCompress() }

// NewBrotliCompress 声明基于 [br] 的压缩算法
//
// [br]: https://www.rfc-editor.org/rfc/rfc7932.html
func NewBrotliCompress(o brotli.WriterOptions) Compressor {
	return compress.NewBrotliCompress(o)
}

// NewLZWCompress 声明基于 lzw 的压缩算法
func NewLZWCompress(order lzw.Order, width int) Compressor {
	return compress.NewLZWCompress(order, width)
}

// NewGzipCompress 声明基于 gzip 的压缩算法
func NewGzipCompress(level int) Compressor { return compress.NewGzipCompress(level) }

// NewDeflateCompress 声明基于 deflate 的压缩算法
func NewDeflateCompress(level int, dict []byte) Compressor {
	return compress.NewDeflateCompress(level, dict)
}

// AllCompresses 所有内置的压缩算法
//
// 可直接供 [Options.Compresses] 使用。
func AllCompresses() []*Compress {
	return []*Compress{
		{Name: "gzip", Compressor: NewGzipCompress(gzip.DefaultCompression)},
		{Name: "deflate", Compressor: NewDeflateCompress(flate.DefaultCompression, nil)},
		{Name: "compress", Compressor: NewLZWCompress(lzw.LSB, 8)},
		{Name: "br", Compressor: NewBrotliCompress(brotli.WriterOptions{})},
		{Name: "zstd", Compressor: NewZstdCompress()},
	}
}
