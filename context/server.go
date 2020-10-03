// SPDX-License-Identifier: MIT

package context

import (
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/v2"
	"github.com/issue9/middleware/v2/compress"
	"github.com/issue9/middleware/v2/debugger"
	"github.com/issue9/middleware/v2/errorhandler"
	"github.com/issue9/mux/v2"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/context/mimetype"
)

// Server 提供了用于构建 Context 对象的基本数据
type Server struct {
	// Interceptor 可以对生成的 Context 在使用前进行修改
	Interceptor func(*Context)

	// Location 指定服务器的时区信息
	//
	// 如果未指定，则会采用 time.Local 作为默认值。
	//
	// 在构建 Context 对象时，该时区信息也会分配给 Context，
	// 如果每个 Context 对象需要不同的值，可以在 Interceptor 中进行修改。
	Location *time.Location

	// Catalog 当前使用的本地化组件
	//
	// 默认情况下会引用 golang.org/x/text/message.DefaultCatalog 对象。
	//
	// golang.org/x/text/message/catalog 提供了 NewBuilder 和 NewFromMap
	// 等方式构建 Catalog 接口实例。
	Catalog catalog.Catalog

	// ResultBuilder 指定生成 Result 数据的方法
	//
	// 默认情况下指向  DefaultResultBuilder。
	ResultBuilder BuildResultFunc

	// middleware
	middlewares   *middleware.Manager
	compress      *compress.Compress
	errorHandlers *errorhandler.ErrorHandler
	debugger      *debugger.Debugger

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
func NewServer(logs *logs.Logs, disableOptions, disableHead bool, root *url.URL) *Server {
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

// AddStatic 添加静态路由
//
// 键名为 URL 的路径部分，相对于项目根路径，键值为文件地址。
//
// 比如在 Domain 和 Root 的值分别为 example.com 和 blog 时，
// 将 Static 的值设置为 /admin ==> ~/data/assets/admin
// 表示将 example.com/blog/admin/* 解析到 ~/data/assets/admin 目录之下。
func (srv *Server) AddStatic(path, dir string) {
	h := http.StripPrefix(path, http.FileServer(http.Dir(dir)))
	srv.Router().Get(path+"{path}", h)
}

// SetErrorHandle 设置指定状态码页面的处理函数
//
// 如果状态码已经存在处理函数，则修改，否则就添加。
// 仅对状态码 >= 400 的有效果。
// 如果 status 为零表示所有未设置的状态码都采用该函数处理。
func (srv *Server) SetErrorHandle(h errorhandler.HandleFunc, status ...int) {
	srv.errorHandlers.Set(h, status...)
}

// AddMiddlewares 设置全局的中间件
//
// 按给定参数的顺序依次调用中间件。
func (srv *Server) AddMiddlewares(middleware ...middleware.Middleware) {
	for _, m := range middleware {
		srv.middlewares.After(m)
	}
}

// SetDebugger 设置调试地址
func (srv *Server) SetDebugger(pprof, vars string) {
	srv.debugger.Pprof = pprof
	srv.debugger.Vars = vars
}

// Handler 将当前服务转换为 http.Handler 接口对象
func (srv *Server) Handler() http.Handler {
	return srv.middlewares
}

// 始终保持这些中间件在最后初始化。用户添加的中间件由 Server.AddMiddlewares 添加。
func (srv *Server) buildMiddlewares() {
	rf := srv.errorHandlers.Recovery(srv.logs.ERROR())
	srv.AddMiddlewares(
		srv.debugger.Middleware, // 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
		rf.Middleware,
		srv.compress.Middleware, // srv.errorhandlers.New 可能会输出大段内容。所以放在其之前。
		srv.errorHandlers.Middleware,
	)
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
