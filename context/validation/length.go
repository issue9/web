// SPDX-License-Identifier: MIT

package validation

import (
	"reflect"
)

type lengthRule struct {
	msg, invalidType string
	min, max         uint64
}

type rangeRule struct {
	msg, invalidType string
	min, max         int64
	valueType        reflect.Type
}

// Length 声明判断内容长度的验证规则
//
// msg 无法验过验证时的信息；
// invalidType 类型无效时的验证信息；
//
// 只能验证类型为 string、Map、Slice 和 Array 的数据。
func Length(msg, invalidType string, min, max uint64) Ruler {
	return &lengthRule{
		msg:         msg,
		invalidType: invalidType,
		min:         min,
		max:         max,
	}
}

// Range 声明判断数值大小的验证规则
//
// 只能验证类型为 int、int8、int16、int32、int64、uint、uint8、uint16、uint32、uint64、float32 和 float64 类型的值。
func Range(msg, invalidType string, min, max int64) Ruler {
	return &rangeRule{
		msg:         msg,
		invalidType: invalidType,
		min:         min,
		max:         max,
		valueType:   reflect.TypeOf(min),
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

	if uint64(l) < rule.min || uint64(l) > rule.max {
		msg = rule.msg
	}
	return
}

func (rule *rangeRule) Validate(v interface{}) (msg string) {
	var ok bool
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(v).Int()
		ok = i >= rule.min && i <= rule.max
	case float32, float64:
		f := reflect.ValueOf(v).Float()
		ok = f >= float64(rule.min) && f <= float64(rule.max)
	default:
		return rule.invalidType
	}

	if !ok {
		msg = rule.msg
	}
	return
}
