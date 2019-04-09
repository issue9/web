// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/language"
	xmessage "golang.org/x/text/message"

	"github.com/issue9/web/internal/resulttest"
)

var (
	_ Result        = &resulttest.Result{}
	_ GetResultFunc = getResult
)

func getResult(status, code int, message string) Result {
	return resulttest.New(status, code, message)
}

func TestApp_Messages(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	a.NotNil(app)

	a.NotError(xmessage.SetString(language.Und, "lang", "und"))
	a.NotError(xmessage.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(xmessage.SetString(language.TraditionalChinese, "lang", "hant"))
	a.NotPanic(func() { (app.AddMessages(400, map[int]string{40010: "lang"})) })

	rslt := app.NewResult(40010)
	r, ok := rslt.(*resulttest.Result)
	a.True(ok).NotNil(r)
	a.Equal(r.Message, "lang").
		Equal(r.Status(), 400)

	// 不存在
	a.Panic(func() {
		app.NewResult(40010001)
	})

	lmsgs := app.Messages(xmessage.NewPrinter(language.Und))
	a.Equal(lmsgs[40010], "und")

	lmsgs = app.Messages(xmessage.NewPrinter(language.SimplifiedChinese))
	a.Equal(lmsgs[40010], "hans")

	lmsgs = app.Messages(xmessage.NewPrinter(language.TraditionalChinese))
	a.Equal(lmsgs[40010], "hant")

	lmsgs = app.Messages(xmessage.NewPrinter(language.English))
	a.Equal(lmsgs[40010], "und")

	lmsgs = app.Messages(nil)
	a.Equal(lmsgs[40010], "lang")
}

func TestApp_AddMessages(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	a.NotPanic(func() {
		app.AddMessages(400, map[int]string{
			1:   "1",
			100: "100",
		})
	})

	msg, found := app.messages[1]
	a.True(found).
		Equal(msg.status, 400).
		Equal(msg.message, "1")

	msg, found = app.messages[401]
	a.False(found).Nil(msg)

	// 消息不能为空
	a.Panic(func() {
		app.AddMessages(400, map[int]string{
			1:   "",
			100: "100",
		})
	})

	// 重复的 ID
	a.Panic(func() {
		app.AddMessages(400, map[int]string{
			1:   "1",
			100: "100",
		})
	})
}
