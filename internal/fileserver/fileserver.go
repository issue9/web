// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package fileserver 封装 http.FileServer 使其可以使用自定义的错误状态码处理。
package fileserver

import (
	"net/http"

	"github.com/issue9/web/internal/exit"
)

type fileServer struct {
	h http.Handler
}

type response struct {
	http.ResponseWriter
}

// New 声明一个可以自定义处理 404 等错误的 FileServer。
//
// 仅对 400 以下的状态作处理。
func New(dir http.Dir) http.Handler {
	return &fileServer{
		h: http.FileServer(dir),
	}
}

func (fs *fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fs.h.ServeHTTP(&response{ResponseWriter: w}, r)
}

func (r *response) WriteHeader(status int) {
	if status >= 400 {
		exit.Context(status)
	}

	r.ResponseWriter.WriteHeader(status)
}
