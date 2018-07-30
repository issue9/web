// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"github.com/issue9/web/context"
	"github.com/issue9/web/module"
	"github.com/issue9/web/result"
)

type (
	// Context 等同于 context.Context，方便调用者使用
	Context = context.Context

	// Result 等同于 result.Result，方便调用者使用
	Result = result.Result

	// Module 等同于 module.Module，方便调用者使用
	Module = module.Module
)
