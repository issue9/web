// SPDX-License-Identifier: MIT

// Package testobject 用于 text 测试对象
package testobject

import (
	"fmt"
	"strconv"
	"strings"
)

// TextObject 本质是实现了 encoding.TextMarshaler 和 encoding.TextUnmarshaler 接口的实例
type TextObject struct {
	Name string
	Age  int
}

// MarshalText 实现 encoding.TextMarshaler 接口
func (o *TextObject) MarshalText() ([]byte, error) {
	return []byte(o.Name + "," + strconv.Itoa(o.Age)), nil
}

// UnmarshalText 实现 encoding.TextUnmarshaler 接口
func (o *TextObject) UnmarshalText(data []byte) error {
	text := strings.Split(string(data), ",")
	if len(text) != 2 {
		return fmt.Errorf("无法转换:%s", string(data))
	}

	age, err := strconv.Atoi(text[1])
	if err != nil {
		return err
	}
	o.Age = age
	o.Name = text[0]
	return nil
}
