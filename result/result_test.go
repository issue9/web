// SPDX-License-Identifier: MIT

package result

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/language"
	xmessage "golang.org/x/text/message"
)

func TestResults_AddMessages(t *testing.T) {
	a := assert.New(t)
	rslt := NewResults(DefaultResultBuilder)
	a.NotNil(rslt)

	a.NotPanic(func() {
		rslt.AddMessages(400, map[int]string{
			1:   "1",
			100: "100",
		})
	})

	msg, found := rslt.messages[1]
	a.True(found).
		Equal(msg.status, 400).
		Equal(msg.message, "1")

	msg, found = rslt.messages[401]
	a.False(found).Nil(msg)

	// 消息不能为空
	a.Panic(func() {
		rslt.AddMessages(400, map[int]string{
			1:   "",
			100: "100",
		})
	})

	// 重复的 ID
	a.Panic(func() {
		rslt.AddMessages(400, map[int]string{
			1:   "1",
			100: "100",
		})
	})
}

func TestResults_Messages(t *testing.T) {
	a := assert.New(t)
	rslt := NewResults(DefaultResultBuilder)
	a.NotNil(rslt)

	a.NotError(xmessage.SetString(language.Und, "lang", "und"))
	a.NotError(xmessage.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(xmessage.SetString(language.TraditionalChinese, "lang", "hant"))
	a.NotPanic(func() {
		rslt.AddMessages(400, map[int]string{40010: "lang"})
	})

	r := rslt.NewResult(40010)
	rr, ok := r.(*defaultResult)
	a.True(ok).NotNil(rr)
	a.Equal(rr.Message, "lang").
		Equal(rr.Status(), 400)

	// 不存在
	a.Panic(func() {
		a.NotError(rslt.NewResult(40010001))
	})

	lmsgs := rslt.Messages(xmessage.NewPrinter(language.Und))
	a.Equal(lmsgs[40010], "und")

	lmsgs = rslt.Messages(xmessage.NewPrinter(language.SimplifiedChinese))
	a.Equal(lmsgs[40010], "hans")

	lmsgs = rslt.Messages(xmessage.NewPrinter(language.TraditionalChinese))
	a.Equal(lmsgs[40010], "hant")

	lmsgs = rslt.Messages(xmessage.NewPrinter(language.English))
	a.Equal(lmsgs[40010], "und")

	lmsgs = rslt.Messages(nil)
	a.Equal(lmsgs[40010], "lang")
}
