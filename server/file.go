// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"time"

	"github.com/issue9/upload"
)

// ServeFile 提供文件下载
//
// 文件可能提供连续的下载功能，其状态码是未定的，
// 所以提供了一个类似于 Render 的变体专门用于下载功能。
//
// path 指向本地文件的地址；
// headers 额外显示的报头内容。
func (ctx *Context) ServeFile(path string, headers map[string]string) {
	for k, v := range headers {
		ctx.Response.Header().Set(k, v)
	}
	http.ServeFile(ctx.Response, ctx.Request, path)
}

// ServeFileFS 提供文件下载服务
//
// 基于 fs.FS 接口获取 p 指向的文件，其它功能与 Context.ServeFile 相同。
func (ctx *Context) ServeFileFS(f fs.FS, p string, headers map[string]string) {
	data, err := fs.ReadFile(f, p)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			ctx.Exit(http.StatusNotFound)
		} else if errors.Is(err, fs.ErrPermission) {
			ctx.Exit(http.StatusForbidden)
		}

		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	var mod time.Time
	s, err := fs.Stat(f, p)
	if err != nil {
		ctx.Server().Logs().Error(err)
		mod = time.Now()
	} else {
		mod = s.ModTime()
	}

	ctx.ServeContent(bytes.NewReader(data), path.Base(p), mod, headers)
}

// ServeContent 将一块内存中的内容转换为文件提供下载
//
// 文件可能提供连续的下载功能，其状态码是未定的，
// 所以提供了一个类似于 Render 的变体专门用于下载功能。
//
// buf 实现 io.ReadSeeker 接口的内存块；
// name 下载时，显示的文件；
// headers 文件报头内容。
func (ctx *Context) ServeContent(buf io.ReadSeeker, name string, mod time.Time, headers map[string]string) {
	for k, v := range headers {
		ctx.Response.Header().Set(k, v)
	}

	http.ServeContent(ctx.Response, ctx.Request, name, mod, buf)
}

// Upload 执行上传文件的相关操作
//
// 返回的是文件列表
func (ctx *Context) Upload(field string, u *upload.Upload) ([]string, error) {
	return u.Do(field, ctx.Request)
}
