// SPDX-License-Identifier: MIT

package content

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/issue9/upload"
)

// DefaultIndexPage ServeFileFS index 参数的默认值
const DefaultIndexPage = "index.html"

// ServeFS 提供基于 fs.FS 的文件下载服
//
// p 表示文件地址，用户应该保证 p 的正确性；
// 如果 p 是目录，则会自动读 p 目录下的 index 文件，
// 如果 index 为空，则采用 DefaultIndexPage 作为其默认值。
func (ctx *Context) ServeFS(f fs.FS, p, index string, headers map[string]string) error {
	if index == "" {
		index = DefaultIndexPage
	}

	if p == "" {
		p = "."
	}

STAT:
	stat, err := fs.Stat(f, p)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		p = path.Join(p, index)
		goto STAT
	}

	data, err := fs.ReadFile(f, p)
	if err != nil {
		return err
	}
	buf := bytes.NewReader(data)

	ctx.ServeContent(buf, filepath.Base(p), stat.ModTime(), headers)
	return nil
}

// ServeContent 将一块内存中的内容转换为文件提供下载
//
// 功能与 http.ServeContent 相同，提供了可自定义报头的功能。
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
