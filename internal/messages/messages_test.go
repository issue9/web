// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package messages

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/language"
	xmessage "golang.org/x/text/message"
)

func TestMessages(t *testing.T) {
	a := assert.New(t)
	m := New()
	a.NotNil(m)

	a.NotError(xmessage.SetString(language.Und, "lang", "und"))
	a.NotError(xmessage.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(xmessage.SetString(language.TraditionalChinese, "lang", "hant"))
	a.NotPanic(func() { (m.newMessage(400, 40010, "lang")) })

	msg, found := m.Message(40010)
	a.True(found).
		Equal(msg.Message, "lang").
		Equal(msg.Status, 400)

	// 不存在
	msg, found = m.Message(40010001)
	a.False(found).Nil(msg)

	lmsgs := m.Messages(xmessage.NewPrinter(language.Und))
	a.Equal(lmsgs[40010], "und")

	lmsgs = m.Messages(xmessage.NewPrinter(language.SimplifiedChinese))
	a.Equal(lmsgs[40010], "hans")

	lmsgs = m.Messages(xmessage.NewPrinter(language.TraditionalChinese))
	a.Equal(lmsgs[40010], "hant")

	lmsgs = m.Messages(xmessage.NewPrinter(language.English))
	a.Equal(lmsgs[40010], "und")

	lmsgs = m.Messages(nil)
	a.Equal(lmsgs[40010], "lang")
}

func TestNewMessages(t *testing.T) {
	a := assert.New(t)
	m := New()
	a.NotNil(m)

	a.NotPanic(func() {
		m.NewMessages(400, map[int]string{
			1:   "1",
			100: "100",
		})
	})

	msg, found := m.messages[1]
	a.True(found).
		Equal(msg.Status, 400).
		Equal(msg.Message, "1")

	msg, found = m.messages[401]
	a.False(found).Nil(msg)

	// 消息不能为空
	a.Panic(func() {
		m.NewMessages(400, map[int]string{
			1:   "",
			100: "100",
		})
	})

	// 重复的 ID
	a.Panic(func() {
		m.NewMessages(400, map[int]string{
			1:   "1",
			100: "100",
		})
	})
}