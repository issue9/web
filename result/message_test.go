// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"testing"

	"github.com/issue9/assert"
)

// cleanMessage 清空所有消息内容
func cleanMessage() {
	messages = map[int]*message{}
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
	cleanMessage()
	a.Error(NewMessages(map[int]string{
		10000: "100",
		40100: "40100",
		99:    "10000",
		100:   "100",
	}))
	msg, found = messages[10000]
	a.True(found).Equal(msg.message, "100")
	msg, found = messages[100]
	a.False(found).Nil(msg)
}
