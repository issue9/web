// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"github.com/issue9/middleware"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/web/config"
	"github.com/issue9/web/mimetype"
)

// Config 配置项
type Config struct {
	Dir                string
	ErrorHandlers      map[int]ErrorHandler
	Compresses         map[string]compress.WriterFunc
	Middlewares        []middleware.Middleware
	ConfigUnmarshals   map[string]config.UnmarshalFunc
	MimetypeMarshals   map[string]mimetype.MarshalFunc
	MimetypeUnmarshals map[string]mimetype.UnmarshalFunc
}
