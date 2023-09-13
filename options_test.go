// SPDX-License-Identifier: MIT

package web

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
)

func TestSanitizeOptions(t *testing.T) {
	a := assert.New(t, false)

	o, err := sanitizeOptions(nil)
	a.NotError(err).NotNil(o)
	a.Equal(o.Location, time.Local).
		NotNil(o.logs).
		NotNil(o.problems).
		NotNil(o.IDGenerator).
		Equal(o.RequestIDKey, RequestIDKey)
}

func TestNewPrinter(t *testing.T) {
	a := assert.New(t, false)

	c := catalog.NewBuilder()
	c.SetString(language.MustParse("zh-CN"), "k1", "zh-cn")
	c.SetString(language.MustParse("zh-TW"), "k1", "zh-tw")

	p := newPrinter(language.MustParse("cmn-hans"), c)
	a.Equal(p.Sprintf("k1"), "zh-cn")
}
