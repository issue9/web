// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package config 提供了对多种格式配置文件的支持
package config

// UnmarshalFunc 定义了将文本内容解析到对象的函数原型。
type UnmarshalFunc func([]byte, interface{}) error
