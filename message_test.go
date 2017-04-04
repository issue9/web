// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func clearMesages() {
	messages = map[int]message{}
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

	clearMesages()
}

func TestNewMessages(t *testing.T) {
	a := assert.New(t)

	a.NotError(NewMessages(map[int]string{
		100:   "100",
		40100: "40100",
	}))

	msg, err := getMessage(100)
	a.NotError(err).NotNil(msg).Equal(msg.status, 100)

	msg, err = getMessage(40100)
	a.NotError(err).NotNil(msg).Equal(msg.status, 401)

	// 不存在
	msg, err = getMessage(100001)
	a.Error(err).NotNil(msg).Equal(msg.status, http.StatusInternalServerError)
}
