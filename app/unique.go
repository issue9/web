// SPDX-License-Identifier: MIT

package app

import (
	"github.com/issue9/unique/v2"

	"github.com/issue9/web"
)

var idGeneratorFactory = map[string]IDGeneratorBuilder{}

// IDGeneratorBuilder 构建生成唯一 ID 的方法
//
// f 表示生成唯一 ID 的方法；
// 表示如果 f 的返回是依赖服务的，那么 srv 即为该服务。
// 否则 srv 为空。
type IDGeneratorBuilder = func() (f web.IDGenerator, srv web.Service)

// RegisterIDGenerator 注册唯一 ID 生成器
//
// 如果同名会被覆盖。
func RegisterIDGenerator(id string, b IDGeneratorBuilder) {
	idGeneratorFactory[id] = b
}

func init() {
	RegisterIDGenerator("date", func() (web.IDGenerator, web.Service) {
		u := unique.NewDate(1000)
		return u.String, u
	})

	RegisterIDGenerator("string", func() (web.IDGenerator, web.Service) {
		u := unique.NewString(1000)
		return u.String, u
	})

	RegisterIDGenerator("number", func() (web.IDGenerator, web.Service) {
		u := unique.NewNumber(1000)
		return u.String, u
	})
}
