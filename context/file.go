// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/issue9/upload"
	"github.com/issue9/utils"

	"github.com/issue9/web/internal/fileserver"
)

// ServeFile 提供文件下载
//
// 文件可能提供连续的下载功能，其状态码是未定的，
// 所以提供了一个类似于 Render 的变体专门用于下载功能。
//
// path 指向本地文件的地址；
// name 下载时，显示的文件，若为空，则直接使用 path 中的文件名部分；
// headers 额外显示的报头内容。
func (ctx *Context) ServeFile(path, name string, headers map[string]string) {
	if !utils.FileExists(path) {
		ctx.Exit(http.StatusNotFound)
	}

	if name == "" {
		name = filepath.Base(path)
	}

	for k, v := range headers {
		ctx.Response.Header().Set(k, v)
	}
	fileserver.ServeFile(ctx.Response, ctx.Request, path)
}

// ServeFileBuffer 将一块内存中的内容转换为文件提供下载
//
// 文件可能提供连续的下载功能，其状态码是未定的，
// 所以提供了一个类似于 Render 的变体专门用于下载功能。
//
// buf 实现 io.ReadSeeker 接口的内存块；
// name 下载时，显示的文件；
// headers 文件报头内容。
func (ctx *Context) ServeFileBuffer(buf io.ReadSeeker, name string, headers map[string]string) {
	for k, v := range headers {
		ctx.Response.Header().Set(k, v)
	}

	fileserver.ServeContent(ctx.Response, ctx.Request, name, time.Now(), buf)
}

// Upload 执行上传文件的相关操作。
//
// 返回的是文件列表
func (ctx *Context) Upload(field string, u *upload.Upload) ([]string, error) {
	return u.Do(field, ctx.Request)
}
