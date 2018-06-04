// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"sort"
	"strconv"
	"strings"
)

// Accept 表示 Accept* 的报头元素
type Accept struct {
	Content string // 完整的内容

	// 解析之后的内容
	Value string
	Q     float32
}

// 将 Content 的内容解析到 Value 和 Q 中
func (a *Accept) parse() error {
	index := strings.IndexByte(a.Content, ';')
	if index < 0 {
		a.Value = a.Content
		a.Q = 1.0
		return nil
	}

	if index >= 0 {
		a.Value = a.Content[:index]
	}

	index = strings.LastIndex(a.Content, ";q=")
	if index > 0 {
		q, err := strconv.ParseFloat(a.Content[index+4:], 32)
		if err != nil {
			return err
		}
		a.Q = float32(q)
	} else {
		a.Q = 1.0
	}

	return nil
}

// AcceptMimeType 从 header 解析出当前请求所需要的解 mimetype 名称和对应的解码函数
//
// 不存在时，返回默认值，出错时，返回错误。
func AcceptMimeType(header string) (string, MarshalFunc, error) {
	accepts, err := ParseAccept(header)
	if err != nil {
		return "", nil, err
	}

	for _, accept := range accepts {
		if m := Marshal(accept.Value); m != nil {
			return accept.Value, m, nil
		}
	}

	return "", nil, ErrUnsupportedMarshal
}

// ParseAccept 将报头内容解析为 []*Accept
func ParseAccept(header string) ([]*Accept, error) {
	accepts := make([]*Accept, 0, strings.Count(header, ",")+1)

	for {
		index := strings.IndexByte(header, ',')
		if index == -1 {
			if header != "" {
				accepts = append(accepts, &Accept{Content: header})
			}
			break
		}

		if index == 0 {
			header = header[1:]
			continue
		}

		val := header[:index]
		if val != "" {
			accepts = append(accepts, &Accept{Content: header[:index]})
		}

		header = header[index+1:]
	}

	for _, accept := range accepts {
		if err := accept.parse(); err != nil {
			return nil, err
		}
	}

	sort.SliceStable(accepts, func(i, j int) bool {
		return accepts[i].Q > accepts[j].Q
	})

	return accepts, nil
}
