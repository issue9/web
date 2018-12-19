// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package exit

import (
	"io"
	"net/http"
	"time"
)

type fileServer struct {
	h http.Handler
}

type response struct {
	http.ResponseWriter
}

// FileServer 声明一个可以自定义处理 404 等错误的 FileServer。
//
// 仅对 400 以下的状态作处理。
func FileServer(dir http.Dir) http.Handler {
	return &fileServer{
		h: http.FileServer(dir),
	}
}

// ServeFile 简单包装 http.ServeFile，使其可以自定义错误状态码。
func ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(&response{ResponseWriter: w}, r, name)
}

// ServeContent 简单包装 http.ServeContent，使其可以自定义错误状态码。
func ServeContent(w http.ResponseWriter, r *http.Request, name string, modified time.Time, buf io.ReadSeeker) {
	http.ServeContent(&response{ResponseWriter: w}, r, name, modified, buf)
}

func (fs *fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fs.h.ServeHTTP(&response{ResponseWriter: w}, r)
}

func (r *response) WriteHeader(status int) {
	if status >= 400 {
		Context(status)
	}

	r.ResponseWriter.WriteHeader(status)
}
