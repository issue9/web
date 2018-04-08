// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"testing"

	"github.com/issue9/assert"
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	r := New(-2, nil) // 不存在的代码
	a.Equal(r.Code, -1)

	code := 400 * 1000
	a.NotError(NewMessage(code, "400"))
	r = New(code, nil)
	a.Equal(r.Message, "400").Equal(r.status, 400).Equal(r.Code, code)

	r = New(code, map[string]string{"f1": "m1", "f2": "m2"})
	a.Equal(r.Message, "400").Equal(r.status, 400).Equal(r.Code, code)
	a.Equal(len(r.Detail), 2)

	cleanMessage()
}

func TestResult_Add_HasDetail(t *testing.T) {
	a := assert.New(t)

	code := 400 * 1000
	a.NotError(NewMessage(code, "400"))
	r := New(code, nil)
	a.False(r.HasDetail())

	r.Add("field", "message")
	r.Add("field", "message")
	a.True(r.HasDetail())
	a.Equal(len(r.Detail), 2)

	cleanMessage()
}

func TestResult_SetDetail(t *testing.T) {
	a := assert.New(t)

	code := 400 * 1000
	a.NotError(NewMessage(code, "400"))
	r := New(code, nil)
	a.False(r.HasDetail())

	r.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
	r.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
	a.True(r.HasDetail())
	a.Equal(len(r.Detail), 4)

	cleanMessage()
}
