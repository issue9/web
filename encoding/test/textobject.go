// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package test 为 encoding 包提供的一些测试功能。
package test

import (
	"encoding"
	"errors"
	"strconv"
	"strings"
)

var (
	_ encoding.TextMarshaler   = &TextObject{}
	_ encoding.TextUnmarshaler = &TextObject{}
)

// TextObject 分别实现了对 encoding.Marshal 和 encoding.Unmarshal 的解析
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
		return errors.New("无法转换")
	}

	age, err := strconv.Atoi(text[1])
	if err != nil {
		return err
	}
	o.Age = age
	o.Name = text[0]
	return nil
}
