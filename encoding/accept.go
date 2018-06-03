// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"strconv"
	"strings"
)

// Accept 表示 Accept* 的报头元素
type Accept struct {
	Value string
	Q     float64
}

// ParseAccept 将报头内容解析为 []*Accept
func ParseAccept(header string) ([]*Accept, error) {
	accepts := make([]*Accept, 0, strings.Count(header, ",")+1)

	for {
		index := strings.IndexByte(header, ',')
		if index == -1 {
			if header != "" {
				accepts = append(accepts, &Accept{Value: header})
			}
			break
		}

		if index == 0 {
			header = header[1:]
			continue
		}

		val := header[:index]
		if val != "" {
			accepts = append(accepts, &Accept{Value: header[:index]})
		}

		header = header[index+1:]
	}

	for _, accept := range accepts {
		val := accept.Value

		index := strings.IndexByte(val, ';')
		if index > 0 {
			accept.Value = val[:index]
		}

		index = strings.LastIndexByte(val, ';')
		if index > 0 {
			q, err := strconv.ParseFloat(val[index+1:], 32)
			if err != nil {
				return nil, err
			}
			accept.Q = q
		}
	}

	return accepts, nil
}
