// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package core 服务于框架本身。
package core

import "net/http"

// Middleware 将一个 http.Handler 封装成另一个 http.Handler
type Middleware func(http.Handler) http.Handler
