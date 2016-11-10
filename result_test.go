// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestNewResult(t *testing.T) {
	a := assert.New(t)

	r := NewResult(-2)
	a.Equal(r.Message, CodeNotExists)

	code := http.StatusBadRequest * Scale
	SetMessage(code, "400")
	r = NewResult(code)
	a.Equal(r.Message, "400")
}

func TestResult_Add_HasDetail(t *testing.T) {
	a := assert.New(t)

	r := NewResult(400 * Scale)
	a.False(r.HasDetail())

	r.Add("field", "message")
	a.True(r.HasDetail())
}

func TestResult_IsError(t *testing.T) {
	a := assert.New(t)

	r := NewResult(400*Scale + 500)
	a.True(r.IsError())

	r = NewResult(300*Scale + 3)
	a.False(r.IsError())
}

func TestResultMarshal(t *testing.T) {
	a := assert.New(t)

	r := NewResult(400)
	r.Message = "400"
	r.Add("field", "message1")
	r.Add("field", "message2")

	bs, err := json.Marshal(r)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":400,"detail":{"field":"message2"}}`)
}

func TestMessage(t *testing.T) {
	a := assert.New(t)

	a.Equal(Message(-1000), CodeNotExists)
}
