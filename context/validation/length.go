// SPDX-License-Identifier: MIT

package validation

import (
	"reflect"
	"strconv"
)

type lengthRule struct {
	msg, invalidType string

	min, max int
	base     int
}

func Length(msg, invalidType string, min, max int, base ...int) Ruler {
	return &lengthRule{
		msg:         msg,
		invalidType: invalidType,
		min:         min,
		max:         max,
		base:        base[0],
	}
}

func (rule *lengthRule) Validate(v interface{}) (msg string) {
	var l int

	switch vv := v.(type) {
	case int:
		l = len(strconv.Itoa(vv))
	case int8:
		l = len(strconv.FormatInt(int64(vv), rule.base))
	case int16:
		l = len(strconv.FormatInt(int64(vv), rule.base))
	case int32:
		l = len(strconv.FormatInt(int64(vv), rule.base))
	case int64:
		l = len(strconv.FormatInt(int64(vv), rule.base))
	case uint:
		l = len(strconv.FormatInt(int64(vv), rule.base))
	case uint8:
		l = len(strconv.FormatInt(int64(vv), rule.base))
	case uint16:
		l = len(strconv.FormatInt(int64(vv), rule.base))
	case uint32:
		l = len(strconv.FormatInt(int64(vv), rule.base))
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
