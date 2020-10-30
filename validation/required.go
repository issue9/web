// SPDX-License-Identifier: MIT

package validation

import "github.com/issue9/is"

// Required 判断值是否必须为非空的规则
//
// skipNil 表示当前值为指针时，如果指向 nil，是否跳过非空检测规则。
// 如果 skipNil 为 false，则 nil 被当作空值处理。
//
// 具体判断规则可参考 github.com/issue9/is.Empty
func Required(msg string, skipNil bool) Ruler {
	return RuleFunc(func(v interface{}) (ret string) {
		if skipNil && v == nil {
			return
		}

		if is.Empty(v, false) {
			ret = msg
		}
		return
	})
}
