// SPDX-FileCopyrightText: 2022-2026 caixw
//
// SPDX-License-Identifier: MIT

package validators

import (
	"net/mail"
	"net/netip"
	"net/url"
	"strings"
	"unicode"
	"uuid"
)

func URL(val string) bool {
	_, err := url.ParseRequestURI(val)
	return err == nil
}

func Email(val string) bool {
	_, err := mail.ParseAddress(val)
	return err == nil
}

func IP4(val string) bool {
	ip, err := netip.ParseAddr(val)
	return err == nil && ip.Is4()
}

func IP6(val string) bool {
	ip, err := netip.ParseAddr(val)
	return err == nil && ip.Is6()
}

// IP 判断是否 IP 地址
//
// 同时支持 [IP4] 和 [IP6]
func IP(val string) bool {
	_, err := netip.ParseAddr(val)
	return err == nil
}

// HexColor 判断一个字符串是否为合法的 16 进制颜色表示法
func HexColor(val string) bool {
	if (len(val) != 4 && len(val) != 7) || val[0] != '#' {
		return false
	}

	return Hex(val[1:])
}

// Hex 是否符合 16 进制数字
func Hex(val string) bool {
	for _, v := range val {
		switch {
		case '0' <= v && v <= '9':
		case 'a' <= v && v <= 'f':
		case 'A' <= v && v <= 'F':
		default:
			return false
		}
	}
	return true
}

func StartWith(prefix string) func(string) bool {
	return func(s string) bool { return strings.HasPrefix(s, prefix) }
}

func EndWith(suffix string) func(string) bool {
	return func(s string) bool { return strings.HasSuffix(s, suffix) }
}

// ASCII ASCII 码
func ASCII(s string) bool {
	for _, c := range s {
		if c > 127 {
			return false
		}
	}
	return true
}

// Alpha 全部都是英文字符
func Alpha(s string) bool {
	for _, c := range s {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return true
}

// Digit 判断字符串是否都为数字
func Digit(val string) bool {
	for _, c := range val {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// UUID 验证 UUID 格式是否正确
func UUID(val string) bool {
	_, err := uuid.Parse(val)
	return err == nil
}

// Empty 字符串是否为空
func Empty(val string) bool { return val == "" }

// Upper 判断字符串是否全部为大写
func Upper(val string) bool {
	for _, c := range val {
		if unicode.IsLower(c) {
			return false
		}
	}
	return true
}

// Lower 判断字符串是否全部为小写
func Lower(val string) bool {
	for _, c := range val {
		if unicode.IsUpper(c) {
			return false
		}
	}
	return true
}
