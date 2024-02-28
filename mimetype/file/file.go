// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package file 提供与文件上下载相关的功能
package file

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/issue9/mux/v7"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
)

// AttachmentHandler 将 name 作为一个附件提供给客户端
//
// fsys 为文件系统；
// name 表示地址中表示文件名部分的参数名称；
// filename 为显示给客户端的文件名，如果为空，则会取 name 的文件名部分；
// inline 是否在浏览器内打开，主要看浏览器本身是否支持；
func AttachmentHandler(fsys fs.FS, name, filename string, inline bool) web.HandlerFunc {
	if fsys == nil {
		panic("参数 fsys 不能为空")
	}
	if name == "" {
		panic("参数 name 不能为空")
	}

	return func(ctx *web.Context) web.Responser {
		p, found := ctx.Route().Params().Get(name) // 空值也是允许的值
		if !found {
			return ctx.NotFound()
		}
		return Attachment(ctx, fsys, p, filename, inline)
	}
}

// ServeFileHandler 构建静态文件服务对象
//
// fsys 为文件系统；
// name 表示地址中表示文件名部分的参数名称；
// index 表示目录下的默认文件名；
func ServeFileHandler(fsys fs.FS, name, index string) web.HandlerFunc {
	if fsys == nil {
		panic("参数 fsys 不能为空")
	}
	if name == "" {
		panic("参数 name 不能为空")
	}

	return func(ctx *web.Context) web.Responser {
		p, _ := ctx.Route().Params().Get(name) // 空值也是允许的值
		return ServeFile(ctx, fsys, p, index)
	}
}

// ServeFile 提供了静态文件服务
//
// name 表示需要读取的文件名，相对于 fsys；
// index 表示 name 为目录时，默认读取的文件，为空表示 index.html；
func ServeFile(ctx *web.Context, fsys fs.FS, name, index string) web.Responser {
	if err := mux.ServeFile(fsys, name, index, ctx, ctx.Request()); err != nil {
		return ctx.Error(err, "")
	}
	return nil
}

// Attachment 提供附件下载服务
//
// name 为相对于 fsys 的文件名；
// filename 为显示给客户端的文件名，如果为空，则会取 name 的文件名部分；
// inline 是否在浏览器内打开，主要看浏览器本身是否支持；
func Attachment(ctx *web.Context, fsys fs.FS, name, filename string, inline bool) web.Responser {
	if filename != "" && strings.ContainsFunc(filename, func(r rune) bool { return r == '/' || r == filepath.Separator }) {
		panic(fmt.Sprintf("filename: %s 不能包含路径分隔符", filename))
	} else if filename == "" {
		filename = filepath.Base(name)
	}
	filename = url.QueryEscape(filename) // 防止中文乱码

	attach := "attachment"
	if inline {
		attach = "inline"
	}

	file, err := fsys.Open(name)
	if err != nil {
		return ctx.Error(err, "")
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return ctx.Error(err, "")
	}

	// 先读 512 字节用于解析文件类型
	const headerSize = 512
	fh := make([]byte, headerSize)
	n, err := file.Read(fh)
	if err != nil {
		return ctx.Error(err, "")
	}
	fh = fh[:n]

	h := ctx.Header()
	h.Set(header.ContentDisposition, attach+"; filename="+filename)
	h.Set(header.ContentType, http.DetectContentType(fh))
	h.Set(header.ContentLength, strconv.FormatInt(stat.Size(), 10))

	if _, err = ctx.Write(fh); err != nil {
		return ctx.Error(err, "")
	}
	if n == headerSize { // 可能没有读完
		if _, err = io.Copy(ctx, file); err != nil {
			return ctx.Error(err, "")
		}
	}

	return nil
}
