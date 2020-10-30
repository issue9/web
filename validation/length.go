// SPDX-License-Identifier: MIT

package validation

import "reflect"

type lengthRule struct {
	msg      string
	min, max int64
}

// Length 声明判断内容长度的验证规则
//
// msg 无法验过验证时的信息；
// 如果 min 和 max 有值为 -1，表示忽略该值的比较，都为 -1 表示不限制长度。
//
// 只能验证类型为 string、Map、Slice 和 Array 的数据。
func Length(msg string, min, max int64) Ruler {
	return &lengthRule{
		msg: msg,
		min: min,
		max: max,
	}
}

// MinLength 声明判断内容长度不小于 min 的验证规则
func MinLength(msg string, min int64) Ruler {
	return Length(msg, min, -1)
}

// MaxLength 声明判断内容长度不大于 max 的验证规则
func MaxLength(msg string, max int64) Ruler {
	return Length(msg, -1, max)
}

func (rule *lengthRule) Validate(v interface{}) (msg string) {
	var l int64
	switch vv := v.(type) {
	case string:
		l = int64(len(vv))
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Array, reflect.Map, reflect.Slice:
			l = int64(rv.Len())
		default:
			return rule.msg
		}
	}

	if rule.min < 0 && rule.max < 0 {
		return ""
	}

	var ok bool
	if rule.min < 0 {
		ok = l <= rule.max // min 已经 i<=0，那么 max 必定 >=0
	} else {
		if ok = l >= rule.min; ok && rule.max > 0 {
			ok = l <= rule.max
		}
	}

	if !ok {
		msg = rule.msg
	}
	return
}
