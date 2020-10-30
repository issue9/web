// SPDX-License-Identifier: MIT

package validation

import (
	"testing"

	"github.com/issue9/assert"
)

func TestIfExpr(t *testing.T) {
	a := assert.New(t)

	count := 5
	rules := If(count == 0, Min("min-5", 5)).Else(Max("max-100", 100), Min("min-50", 50)).Rules()
	a.Equal(2, len(rules))

	// 第二次 Else 清空了之前的规则
	count = 5
	rules = If(count == 0, Min("min-5", 5)).Else(Max("max-100", 100), Min("min-50", 50)).Else().Rules()
	a.Equal(0, len(rules))

	count = 0
	rules = If(count == 0, Min("min-5", 5)).Else(Max("max-100", 100), Min("min-50", 50)).Rules()
	a.Equal(1, len(rules))
}
