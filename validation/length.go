// SPDX-License-Identifier: MIT

package validation

import "reflect"

type lengthRule struct {
	msg, invalidType string
	min, max         int64
}

type rangeRule lengthRule

// Length 声明判断内容长度的验证规则
//
// msg 无法验过验证时的信息；
// invalidType 类型无效时的验证信息；
//
// 只能验证类型为 string、Map、Slice 和 Array 的数据。
func Length(msg, invalidType string, min, max int64) Ruler {
	return newLength(msg, invalidType, min, max)
}

// MinLength 声明判断内容长度不小于 min 的验证规则
func MinLength(msg, invalidType string, min int64) Ruler {
	return Length(msg, invalidType, min, -1)
}

// MaxLength 声明判断内容长度不大于 max 的验证规则
func MaxLength(msg, invalidType string, max int64) Ruler {
	return Length(msg, invalidType, -1, max)
}

// Range 声明判断数值大小的验证规则
//
// 只能验证类型为 int、int8、int16、int32、int64、uint、uint8、uint16、uint32、uint64、float32 和 float64 类型的值。
func Range(msg, invalidType string, min, max int64) Ruler {
	return (*rangeRule)(newLength(msg, invalidType, min, max))
}

// Min 声明判断数值不小于 min 的验证规则
func Min(msg, invalidType string, min int64) Ruler {
	return Range(msg, invalidType, min, -1)
}

// Max 声明判断数值不大于 max 的验证规则
func Max(msg, invalidType string, max int64) Ruler {
	return Range(msg, invalidType, -1, max)
}

func newLength(msg, invalidType string, min, max int64) *lengthRule {
	if min == -1 && max == -1 {
		panic("min 和 max 不能同时为 -1")
	}

	return &lengthRule{
		msg:         msg,
		invalidType: invalidType,
		min:         min,
		max:         max,
	}
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
			return rule.invalidType
		}
	}

	var ok bool
	if rule.min < 0 {
		ok = l <= rule.max // min 已经小于 0，那么 max 必定大于 0
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

func (rule *rangeRule) Validate(v interface{}) (msg string) {
	var l int64
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		l = reflect.ValueOf(v).Int()
	case float32, float64:
		l = int64(reflect.ValueOf(v).Float())
	default:
		return rule.invalidType
	}

	var ok bool
	if rule.min < 0 {
		ok = l <= rule.max // min 已经小于 0，那么 max 必定大于 0
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
