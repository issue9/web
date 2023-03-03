// SPDX-License-Identifier: MIT

package app

import (
	"github.com/issue9/unique/v2"

	"github.com/issue9/web/server"
)

var uniqueGeneratorFactory = map[string]UniqueGeneratorBuilder{}

type UniqueGeneratorBuilder func() server.UniqueGenerator

// RegisterUniqueGenerator 注册唯一 ID 生成器
//
// 如果同名会被覆盖。
func RegisterUniqueGenerator(id string, b UniqueGeneratorBuilder) {
	uniqueGeneratorFactory[id] = b
}

func init() {
	RegisterUniqueGenerator("date", func() server.UniqueGenerator {
		return unique.NewDate(1000)
	})

	RegisterUniqueGenerator("string", func() server.UniqueGenerator {
		return unique.NewString(1000)
	})

	RegisterUniqueGenerator("number", func() server.UniqueGenerator {
		return unique.NewNumber(1000)
	})
}
