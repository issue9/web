// SPDX-License-Identifier: MIT

package utils

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestIsEmail(t *testing.T) {
	a := assert.New(t, false)
	a.True(IsEmail("abc@example.com"))
	a.False(IsEmail("@example.com"))
	a.False(IsEmail("https://example.com"))
	a.False(IsEmail("example.com"))
}

func TestIsURL(t *testing.T) {
	a := assert.New(t, false)
	a.True(IsURL("https://example.com"))
	a.True(IsURL("http://example.com"))
	a.False(IsURL("ftp://example.com"))
}

func TestCutTag(t *testing.T) {
	a := assert.New(t, false)

	tag, suffix := CutTag("@tag tag desc ")
	a.Equal(tag, "@tag").
		Equal(suffix, "tag desc")

	tag, suffix = CutTag("@tag   ")
	a.Equal(tag, "@tag").
		Equal(suffix, "")

	tag, suffix = CutTag("@tag")
	a.Equal(tag, "@tag").
		Equal(suffix, "")
}

func TestSplitSpaceN(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		SplitSpaceN("", 0)
	}, "参数 maxSize 不能为 0")

	s, l := SplitSpaceN("", 1)
	a.Equal(l, 0).Nil(s)

	s, l = SplitSpaceN("    ", 1)
	a.Equal(l, 0).Nil(s)

	s, l = SplitSpaceN("ab1", 1)
	a.Equal(l, 1).Equal(s, []string{"ab1"})

	s, l = SplitSpaceN("ab3", 3)
	a.Equal(l, 1).Equal(s, []string{"ab3", "", ""})

	s, l = SplitSpaceN("  ab1  ", 1)
	a.Equal(l, 1).Equal(s, []string{"ab1"})

	s, l = SplitSpaceN("  a\tb1  ", 1)
	a.Equal(l, 1).Equal(s, []string{"a\tb1"})

	s, l = SplitSpaceN("  a\tb2  ", 2)
	a.Equal(l, 2).Equal(s, []string{"a", "b2"})

	s, l = SplitSpaceN("  a\tb3  ", 3)
	a.Equal(l, 2).Equal(s, []string{"a", "b3", ""})

	s, l = SplitSpaceN("  aa\tbb1  ", 1)
	a.Equal(l, 1).Equal(s, []string{"aa\tbb1"})

	s, l = SplitSpaceN("  aa\tbb2  ", 2)
	a.Equal(l, 2).Equal(s, []string{"aa", "bb2"})

	s, l = SplitSpaceN("  aa\tbb3  ", 3)
	a.Equal(l, 2).Equal(s, []string{"aa", "bb3", ""})

	s, l = SplitSpaceN("  aa\t  \t bb3  ", 3)
	a.Equal(l, 2).Equal(s, []string{"aa", "bb3", ""})

	s, l = SplitSpaceN("  aa\t  \t bb4  ", 4)
	a.Equal(l, 2).Equal(s, []string{"aa", "bb4", "", ""})
}
