// SPDX-License-Identifier: MIT

package context

import (
	"net/url"
	"path"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/v2"
	"github.com/issue9/middleware/v2/compress"
	"github.com/issue9/middleware/v2/debugger"
	"github.com/issue9/middleware/v2/errorhandler"
	"github.com/issue9/mux/v3"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/context/mimetype"
)

// Server 提供了用于构建 Context 对象的基本数据
type Server struct {
	// Location 指定服务器的时区信息
	//
	// 如果未指定，则会采用 time.Local 作为默认值。
	//
	// 在构建 Context 对象时，该时区信息也会分配给 Context，
	// 如果每个 Context 对象需要不同的值，可以通过 AddFilters 进行修改。
	Location *time.Location

	// Catalog 当前使用的本地化组件
	//
	// 默认情况下会引用 golang.org/x/text/message.DefaultCatalog 对象。
	//
	// golang.org/x/text/message/catalog 提供了 NewBuilder 和 NewFromMap
	// 等方式构建 Catalog 接口实例。
	//
	// NOTE: Context.LocalePrinter 在初始化时与当前值进行关联。
	// 如果中途修改 Catalog 的值，已经建立的 Context 实例中的
	// LocalePrinter 并不会与新的 Catalog 进行关联，依然指向旧值。
	Catalog catalog.Catalog

	// ResultBuilder 指定生成 Result 数据的方法
	//
	// 默认情况下指向  DefaultResultBuilder。
	ResultBuilder BuildResultFunc

	// 保存 Context 在存续期间的可复用变量
	//
	// 这是比 context.Value 更经济的传递变量方式。
	//
	// 如果仅需要在单次请求中传递参数，可直接使用 Context.Vars。
	Vars map[interface{}]interface{}

	cache cache.Cache

	// middleware
	middlewares   *middleware.Manager
	compress      *compress.Compress
	errorHandlers *errorhandler.ErrorHandler
	debugger      *debugger.Debugger
	filters       []Filter

	// url
	root   string
	url    *url.URL
	router *mux.Prefix

	logs      *logs.Logs
	uptime    time.Time
	mimetypes *mimetype.Mimetypes

	// result
	messages map[int]*resultMessage
}

// NewServer 返回 *Server 实例
func NewServer(logs *logs.Logs, cache cache.Cache, disableOptions, disableHead bool, root *url.URL) *Server {
	// NOTE: Server 中在初始化之后不能修改的都由 NewServer 指定，
	// 其它字段的内容给定一个初始值，后期由用户自行决定。

	// 保证不以 / 结尾
	if len(root.Path) > 0 && root.Path[len(root.Path)-1] == '/' {
		root.Path = root.Path[:len(root.Path)-1]
	}

	mux := mux.New(disableOptions, disableHead, false, nil, nil)
	router := mux.Prefix(root.Path)

	srv := &Server{
		Location:      time.Local,
		Catalog:       message.DefaultCatalog,
		ResultBuilder: DefaultResultBuilder,

		Vars: map[interface{}]interface{}{},

		cache: cache,

		middlewares: middleware.NewManager(router.Mux()),
		compress: compress.New(logs.ERROR(), map[string]compress.WriterFunc{
			"gzip":    compress.NewGzip,
			"deflate": compress.NewDeflate,
			"br":      compress.NewBrotli,
		}, "*"),
		errorHandlers: errorhandler.New(),
		debugger:      &debugger.Debugger{},

		root:   root.String(),
		url:    root,
		router: router,

		logs:      logs,
		uptime:    time.Now(),
		mimetypes: mimetype.NewMimetypes(),

		messages: make(map[int]*resultMessage, 20),
	}

	srv.buildMiddlewares()

	return srv
}

// Logs 返回关联的 logs.Logs 实例
func (srv *Server) Logs() *logs.Logs {
	return srv.logs
}

// Cache 返回缓存的相关接口
func (srv *Server) Cache() cache.Cache {
	return srv.cache
}

// Uptime 当前服务的运行时间
func (srv *Server) Uptime() time.Time {
	return srv.uptime
}

// Now 返回当前时间
//
// 与 time.Now() 的区别在于 Now() 基于当前时区
func (srv *Server) Now() time.Time {
	return time.Now().In(srv.Location)
}

// ParseTime 分析基于当前时区的时间
func (srv *Server) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, srv.Location)
}

// Server 获取关联的 context.Server 实例
func (ctx *Context) Server() *Server {
	return ctx.server
}

// Path 生成路径部分的地址
func (srv *Server) Path(p string) string {
	p = path.Join(srv.url.Path, p)
	if p != "" && p[0] != '/' {
		p = "/" + p
	}

	return p
}

// URL 构建一条基于 Root 的完整 URL
func (srv *Server) URL(p string) string {
	switch {
	case len(p) == 0:
		return srv.root
	case p[0] == '/':
		// 由 NewServer 保证 root 不能 / 结尾
		return srv.root + p
	default:
		return srv.root + "/" + p
	}
}

// AddMarshals 添加多个编码函数
func (srv *Server) AddMarshals(ms map[string]mimetype.MarshalFunc) error {
	return srv.mimetypes.AddMarshals(ms)
}

// AddMarshal 添加编码函数
//
// mf 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP
// 另作处理，比如下载，上传等内容。
func (srv *Server) AddMarshal(name string, mf mimetype.MarshalFunc) error {
	return srv.mimetypes.AddMarshal(name, mf)
}

// AddUnmarshals 添加多个编码函数
func (srv *Server) AddUnmarshals(ms map[string]mimetype.UnmarshalFunc) error {
	return srv.mimetypes.AddUnmarshals(ms)
}

// AddUnmarshal 添加编码函数
//
// mm 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP
// 另作处理，比如下载，上传等内容。
func (srv *Server) AddUnmarshal(name string, mm mimetype.UnmarshalFunc) error {
	return srv.mimetypes.AddUnmarshal(name, mm)
}
