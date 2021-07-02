// SPDX-License-Identifier: MIT

package content

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/language"
)

func TestContent_catalog(t *testing.T) {
	a := assert.New(t)

	c := New(DefaultBuilder)
	a.NotNil(c)
	l := c.LocaleBuilder(language.SimplifiedChinese)
	l.SetString("test", "测试")
	l = c.LocaleBuilder(language.Und)
	l.SetString("test", "und")

	p := c.NewLocalePrinter(language.SimplifiedChinese)
	a.Equal(p.Sprintf("test"), "测试")
	p = c.NewLocalePrinter(language.Japanese)
	a.Equal(p.Sprintf("test"), "und")
}
