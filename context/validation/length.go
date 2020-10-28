// SPDX-License-Identifier: MIT

package validation

import (
	"reflect"
)

type lengthRule struct {
	msg, invalidType string

	min, max int
	base     int
}

// Length 声明判断内容长度的验证规则
//
// msg 无法验过验证时的信息；
// invalidType 类型无效时的验证信息；
//
// 只能验证类型为 string、Map、Slice 和 Array 的数据。
func Length(msg, invalidType string, min, max int) Ruler {
	return &lengthRule{
		msg:         msg,
		invalidType: invalidType,
		min:         min,
		max:         max,
	}
}

func (rule *lengthRule) Validate(v interface{}) (msg string) {
	var l int

	switch vv := v.(type) {
	case string:
		l = len(vv)
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Array, reflect.Map, reflect.Slice:
			l = rv.Len()
		default:
			return rule.invalidType
		}
	}

	if l < rule.min || l > rule.max {
		msg = rule.msg
	}
	return
}
