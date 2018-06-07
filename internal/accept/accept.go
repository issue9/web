// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package accept 用于处理 accpet 系列的报头。
package accept

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
func parseAccept(v string) (val string, q float32, err error) {
	index := strings.IndexByte(v, ';')
	if index < 0 {
		return v, 1.0, nil
	}

	if index >= 0 {
		val = v[:index]
	}

	index = strings.LastIndex(v, ";q=")
	if index >= 0 {
		qq, err := strconv.ParseFloat(v[index+3:], 32)
		if err != nil {
			return "", 0, err
		}
		q = float32(qq)
	} else {
		q = 1.0
	}

	return val, q, nil
}

// Parse 将报头内容解析为 []*Accept
func Parse(header string) ([]*Accept, error) {
	accepts := make([]*Accept, 0, strings.Count(header, ",")+1)

	for {
		index := strings.IndexByte(header, ',')
		if index == -1 {
			if header != "" {
				val, q, err := parseAccept(header)
				if err != nil {
					return nil, err
				}
				if q > 0.0 {
					accepts = append(accepts, &Accept{Content: header, Value: val, Q: q})
				}
			}
			break
		}

		if index == 0 {
			header = header[1:]
			continue
		}

		v := header[:index]
		if v != "" {
			val, q, err := parseAccept(v)
			if err != nil {
				return nil, err
			}

			if q > 0.0 {
				accepts = append(accepts, &Accept{Content: v, Value: val, Q: q})
			}
		}

		header = header[index+1:]
	}

	sort.SliceStable(accepts, func(i, j int) bool {
		return accepts[i].Q > accepts[j].Q
	})

	return accepts, nil
}
