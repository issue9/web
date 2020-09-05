// SPDX-License-Identifier: MIT

package server

import (
	"net/http"

	"github.com/issue9/mux/v2"
)

// Prefix 声明一个 Prefix 实例。
func (srv *Server) Prefix(prefix string) *mux.Prefix {
	return srv.router.Prefix(prefix)
}

// Handle 添加一个路由项
func (srv *Server) Handle(path string, h http.Handler, methods ...string) error {
	return srv.router.Handle(path, h, methods...)
}

// Get 指定一个 GET 请求
func (srv *Server) Get(path string, h http.Handler) *mux.Prefix {
	return srv.router.Get(path, h)
}

// Post 指定个 POST 请求处理
func (srv *Server) Post(path string, h http.Handler) *mux.Prefix {
	return srv.router.Post(path, h)
}

// Delete 指定个 Delete 请求处理
func (srv *Server) Delete(path string, h http.Handler) *mux.Prefix {
	return srv.router.Delete(path, h)
}

// Put 指定个 Put 请求处理
func (srv *Server) Put(path string, h http.Handler) *mux.Prefix {
	return srv.router.Put(path, h)
}

// Patch 指定个 Patch 请求处理
func (srv *Server) Patch(path string, h http.Handler) *mux.Prefix {
	return srv.router.Patch(path, h)
}

// HandleFunc 指定一个请求
func (srv *Server) HandleFunc(path string, h func(w http.ResponseWriter, r *http.Request), methods ...string) error {
	return srv.router.HandleFunc(path, h, methods...)
}

// GetFunc 指定一个 GET 请求
func (srv *Server) GetFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return srv.router.GetFunc(path, h)
}

// PostFunc 指定一个 Post 请求
func (srv *Server) PostFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return srv.router.PostFunc(path, h)
}

// DeleteFunc 指定一个 Delete 请求
func (srv *Server) DeleteFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return srv.router.DeleteFunc(path, h)
}

// PutFunc 指定一个 Put 请求
func (srv *Server) PutFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return srv.router.PutFunc(path, h)
}

// PatchFunc 指定一个 Patch 请求
func (srv *Server) PatchFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return srv.router.PatchFunc(path, h)
}
