// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"errors"

	"github.com/issue9/web/encoding"
)

var (
	// ErrExists 表示存在相同名称的项。
	// 多个类似功能的函数，都有可能返回此错误。
	ErrExists = errors.New("存在相同名称的项")
)

var (
	marshals = map[string]encoding.Marshal{
		encoding.DefaultEncoding: encoding.TextMarshal,
	}

	unmarshals = map[string]encoding.Unmarshal{
		encoding.DefaultEncoding: encoding.TextUnmarshal,
	}

	charset = map[string]encoding.Charset{
		encoding.DefaultCharset: nil,
	}
)

// AddMarshal 添加一个新的解码器，只有通过 AddMarshal 添加的解码器，
// 才能被 Context 使用。
func AddMarshal(name string, m encoding.Marshal) error {
	_, found := marshals[name]
	if found {
		return ErrExists
	}

	marshals[name] = m
	return nil
}

// AddUnmarshal 添加一个编码器，只有通过 AddUnmarshal 添加的解码器，
// 才能被 Context 使用。
func AddUnmarshal(name string, m encoding.Unmarshal) error {
	_, found := unmarshals[name]
	if found {
		return ErrExists
	}

	unmarshals[name] = m
	return nil
}

// AddCharset 添加编码方式，只有通过 AddCharset 添加的字符集，
// 才能被 Context 使用。
func AddCharset(name string, enc encoding.Charset) error {
	_, found := charset[name]
	if found {
		return ErrExists
	}

	charset[name] = enc
	return nil
}
