// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package message

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestGetStatus(t *testing.T) {
	a := assert.New(t)

	a.Equal(getStatus(100), 100)
	a.Equal(getStatus(200), 200)
	a.Equal(getStatus(211), 211)
	a.Equal(getStatus(9011), 901)
	a.Equal(getStatus(9099), 909)
}

func TestRegister(t *testing.T) {
	a := assert.New(t)

	a.Error(Register(99, "99"))      // 必须大于等于 100
	a.NotError(Register(100, "100")) // 必须大于等于 100

	a.Error(Register(100, "100")) // 已经存在
	a.Error(Register(100, ""))    // 消息为空

	Clean()
}

func TestRegisters(t *testing.T) {
	a := assert.New(t)

	a.NotError(Registers(map[int]string{
		100:   "100",
		40100: "40100",
	}))

	msg, err := GetMessage(100)
	a.NotError(err).NotNil(msg).Equal(msg.Status, 100)

	msg, err = GetMessage(40100)
	a.NotError(err).NotNil(msg).Equal(msg.Status, 401)

	// 不存在
	msg, err = GetMessage(100001)
	a.Error(err).NotNil(msg).Equal(msg.Status, http.StatusInternalServerError)
}
