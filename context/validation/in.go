// SPDX-License-Identifier: MIT

package validation

import (
	"reflect"

	"github.com/issue9/sliceutil"
)

type inRule struct {
	elements []interface{}
	msg      string
}

type notInRule inRule

// In 声明枚举类型的验证规则
//
// 要求验证的值必须包含在 element 元素中，如果不存在，则返回 msg 的内容。
func In(msg string, element ...interface{}) Ruler {
	return &inRule{
		msg:      msg,
		elements: element,
	}
}

// NotIn 声明不在枚举中的验证规则
//
// 正好与 In 规则相反。
func NotIn(msg string, element ...interface{}) Ruler {
	return &notInRule{
		msg:      msg,
		elements: element,
	}
}

func (rule *inRule) Validate(v interface{}) (msg string) {
	if !rule.in(v) {
		return rule.msg
	}
	return
}

func (rule *notInRule) Validate(v interface{}) (msg string) {
	if r := (*inRule)(rule); r.in(v) {
		return rule.msg
	}
	return
}

func (rule *inRule) in(v interface{}) bool {
	return sliceutil.Count(rule.elements, func(i int) bool {
		elem := rule.elements[i]

		switch v.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			elemType := reflect.TypeOf(elem)
			rv := reflect.ValueOf(v)

			if !rv.Type().ConvertibleTo(elemType) {
				return false
			}
			return rv.Convert(elemType).Interface() == elem
		default:
			return reflect.DeepEqual(v, rule.elements[i])
		}
	}) > 0
}
