// SPDX-License-Identifier: MIT

// Package web 一个微型的 RESTful API 框架
package web

import (
	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/version"
)

// Version 当前框架的版本
const Version = version.Version

type Server = context.Server

type Context = context.Context
