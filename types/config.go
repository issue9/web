// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package types

// Config 需要写入到 web.json 配置文件的类需要实现的接口。
type Config interface {
	// 对配置内容进行初始化，对一些可以为空的值，进行默认赋值。
	Init() error

	// 检测配置项是否有错误。
	Check() error
}
