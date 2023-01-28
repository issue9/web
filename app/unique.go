// SPDX-License-Identifier: MIT

package app

import (
	"time"

	"github.com/issue9/unique/v2"

	"github.com/issue9/web/internal/service"
)

var uniqueGeneratorFactory = map[string]UniqueGeneratorBuilder{}

// UniqueGenerator 唯一 ID 生成器的接口
type UniqueGenerator interface {
	service.Servicer
	ID() string
}

type UniqueGeneratorBuilder func() UniqueGenerator

// RegisterUniqueGenerator 注册唯一 ID 生成器
//
// 如果同名会被覆盖。
func RegisterUniqueGenerator(id string, b UniqueGeneratorBuilder) {
	uniqueGeneratorFactory[id] = b
}

func init() {
	RegisterUniqueGenerator("date", func() UniqueGenerator {
		return newGenerator("20060102150405-", 10)
	})

	RegisterUniqueGenerator("string", func() UniqueGenerator {
		return newGenerator("", 36)
	})

	RegisterUniqueGenerator("number", func() UniqueGenerator {
		return newGenerator("", 10)
	})
}

type generator struct {
	*unique.Unique
}

func newGenerator(prefixFormat string, base int) UniqueGenerator {
	return &generator{
		Unique: unique.New(1000, time.Hour, prefixFormat, base),
	}
}

func (u *generator) ID() string { return u.String() }
