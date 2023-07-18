// SPDX-License-Identifier: MIT

// Package utils 公用方法
package utils

import (
	"strings"
	"unicode"
)

func IsEmail(s string) bool {
	return !IsURL(s) && strings.IndexByte(s, '@') > 0
}

func IsURL(s string) bool {
	s = strings.ToLower(s)
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// CutTag 从第一个空格处截为两段
func CutTag(line string) (tag, suffix string) {
	words, _ := SplitSpaceN(line, 2)
	if len(words) != 2 {
		panic(line)
	}
	return words[0], words[1]
}

// SplitSpaceN 以空格分隔字符串
//
// maxSize 表示最多分隔的数量，如果无法达到 maxSize 的数量，则采用空字符串代替剩余的元素，
// 返回值 length 表示实际的元素数量。-1 表示按实际的数量拆分，length 始终等于 len(ret)。
func SplitSpaceN(s string, maxSize int) (ret []string, length int) {
	if maxSize == 0 {
		panic("参数 maxSize 不能为 0")
	} else if maxSize == 1 {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil, 0
		}
		return []string{s}, 1
	} else if maxSize < 0 {
		ret = make([]string, 0, 10)
	} else {
		ret = make([]string, 0, maxSize)
	}

	var prevIndex int
	prevIsSpace := true
	count := -1
	maxSize--

	for index, c := range s {
		if !unicode.IsSpace(c) {
			if !prevIsSpace {
				continue
			}

			prevIndex = index
			prevIsSpace = false
			count++

			if maxSize > 0 && count == maxSize { // maxSize <= 1 的情况在开头已经处理过。
				break
			}
			continue
		}

		if !prevIsSpace { // 连续空格中的第一个空格
			prevIsSpace = true
			ret = append(ret, s[prevIndex:index])
			prevIndex = index
		}
	}

	if last := strings.TrimRightFunc(s[prevIndex:], unicode.IsSpace); last != "" {
		ret = append(ret, last)
	}

	l := len(ret)
	if maxSize >= 0 { // 填充空白
		maxSize++
		for maxSize-len(ret) > 0 {
			ret = append(ret, "")
		}
	}

	return ret, l
}

// SplitSpace 以空格分隔字符串
func SplitSpace(s string) []string {
	ret, _ := SplitSpaceN(s, -1)
	return ret
}
