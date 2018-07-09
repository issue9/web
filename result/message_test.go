// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/language"
	xmessage "golang.org/x/text/message"
)

// cleanMessage 清空所有消息内容
func cleanMessage() {
	messages = map[int]*message{}
}

func TestMessages(t *testing.T) {
	a := assert.New(t)

	a.NotError(xmessage.SetString(language.Und, "lang", "und"))
	a.NotError(xmessage.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(xmessage.SetString(language.TraditionalChinese, "lang", "hant"))
	a.NotError(NewMessage(40010, "lang"))

	lmsgs := LocaleMessages(xmessage.NewPrinter(language.Und))
	msgs := Messages()
	a.Equal(lmsgs[40010], "und")

	lmsgs = LocaleMessages(xmessage.NewPrinter(language.SimplifiedChinese))
	msgs = Messages()
	a.Equal(lmsgs[40010], "hans")
	a.Equal(msgs[40010], "lang")

	lmsgs = LocaleMessages(xmessage.NewPrinter(language.TraditionalChinese))
	msgs = Messages()
	a.Equal(lmsgs[40010], "hant")
	a.Equal(msgs[40010], "lang")

	lmsgs = LocaleMessages(xmessage.NewPrinter(language.English))
	msgs = Messages()
	a.Equal(lmsgs[40010], "und")
	a.Equal(msgs[40010], "lang")

	cleanMessage()
}

func TestFindMessage(t *testing.T) {
	a := assert.New(t)

	a.NotError(NewMessage(10010, "100"))

	msg := findMessage(10010)
	a.Equal(msg.status, 100).
		Equal(msg.message, "100")

	msg = findMessage(100) // 不存在
	a.Equal(msg, unknownCodeMessage)

	cleanMessage()
}

func TestGetStatus(t *testing.T) {
	a := assert.New(t)

	a.Equal(getStatus(100), 100)
	a.Equal(getStatus(200), 200)
	a.Equal(getStatus(211), 211)
	a.Equal(getStatus(9011), 901)
	a.Equal(getStatus(9099), 909)
}

func TestNewMessage(t *testing.T) {
	a := assert.New(t)

	a.Error(NewMessage(99, "99"))      // 必须大于等于 100
	a.NotError(NewMessage(100, "100")) // 必须大于等于 100

	a.Error(NewMessage(100, "100")) // 已经存在
	a.Error(NewMessage(100, ""))    // 消息为空

	cleanMessage()
}

func TestNewMessages(t *testing.T) {
	a := assert.New(t)

	a.NotError(NewMessages(map[int]string{
		100:   "100",
		40100: "40100",
	}))

	msg, found := messages[100]
	a.True(found).
		NotNil(msg).
		Equal(msg.status, 100)
	a.Equal(len(Messages()), 2)

	msg, found = messages[40100]
	a.True(found).
		NotNil(msg).
		Equal(msg.status, 401)

	// 不存在
	msg, found = messages[100001]
	a.False(found).Nil(msg)

	// 小于 100 的值，会发生错误
	a.Error(NewMessages(map[int]string{
		10000: "100",
		40100: "40100",
		99:    "10000",
		100:   "100",
	}))

	cleanMessage()
}

func TestNewStatusMessage(t *testing.T) {
	a := assert.New(t)
	a.NotError(NewStatusMessage(500, 50010, "100"))

	a.Error(NewStatusMessage(500, UnknownCode, "msg")) // 错误代码不正确
	a.Error(NewStatusMessage(500, 50010, ""))          // msg 不能为空
	a.Error(NewStatusMessage(500, 50010, "100"))       // 已经存在
	a.Error(NewStatusMessage(600, 50010, "100"))       // 已经存在，仅使状态码不同

	cleanMessage()
}

func TestNewStatusMessages(t *testing.T) {
	a := assert.New(t)

	a.NotError(NewStatusMessages(400, map[int]string{
		1:   "1",
		100: "100",
	}))

	msg, found := messages[1]
	a.True(found).
		Equal(msg.status, 400).
		Equal(msg.message, "1")

	msg, found = messages[401]
	a.False(found).Nil(msg)

	a.Error(NewStatusMessages(400, map[int]string{
		1:   "",
		100: "100",
	}))

	cleanMessage()
}
