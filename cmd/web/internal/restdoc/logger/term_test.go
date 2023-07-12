// SPDX-License-Identifier: MIT

package logger

import (
	"bytes"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

func TestTypeColors(t *testing.T) {
	a := assert.New(t, false)
	a.Equal(typeSize, len(typeColors))
}

func TestBuildTermHandler(t *testing.T) {
	a := assert.New(t, false)

	cata := catalog.NewBuilder()
	cata.SetString(language.SimplifiedChinese, "%s at %s:%d\n", "%s 位于 %s:%d\n")
	cata.SetString(language.SimplifiedChinese, "error", "ERROR")
	p := message.NewPrinter(language.SimplifiedChinese, message.Catalog(cata))
	buf := new(bytes.Buffer)
	l := New(BuildTermHandler(buf, p))
	a.NotNil(l)

	l.Log(Info, "error", "f.go", 10)
	a.Contains(buf.String(), "error 位于 f.go:10\n")

	buf.Reset()
	l.Log(Info, localeutil.Error("error"), "f.go", 10)
	a.Contains(buf.String(), "ERROR 位于 f.go:10\n")
}
