// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package texttest 包含了一个实现 text 解码的对象，方便其它包测试。
package texttest

import (
	"fmt"
	"strconv"
	"strings"
)

// TextObject 本质是实现了 encoding.TextMarshaler 和 encoding.TextUnmarshaler
// 接口的实例。
//
// 可用于 encoding 包中的 MarshalFunc 和 UnmarshalFunc 的测试。
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
