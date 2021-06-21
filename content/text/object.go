// SPDX-License-Identifier: MIT

package text

import (
	"fmt"
	"strconv"
	"strings"
)

// TestObject 本质是实现了 encoding.TextMarshaler 和 encoding.TextUnmarshaler
// 接口的实例。
//
// 可用于 mimetype 包中的 MarshalFunc 和 UnmarshalFunc 的测试。
type TestObject struct {
	Name string
	Age  int
}

// MarshalText 实现 encoding.TextMarshaler 接口
func (o *TestObject) MarshalText() ([]byte, error) {
	return []byte(o.Name + "," + strconv.Itoa(o.Age)), nil
}

// UnmarshalText 实现 encoding.TextUnmarshaler 接口
func (o *TestObject) UnmarshalText(data []byte) error {
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
