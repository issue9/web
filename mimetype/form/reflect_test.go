// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package form

import (
	"testing"

	"github.com/issue9/assert"
)

type tagObject struct {
	Name     string
	Age      int      `form:"age"`
	Friend   []string `form:"friend"`
	Ignore   string   `form:"-"`
	unexport bool
}

var formTagString = "Name=Ava&age=10&friend=Jess&friend=Sarah&friend=Zoe"

var tagObjectData = &tagObject{
	Name:     "Ava",
	Age:      10,
	Friend:   []string{"Jess", "Sarah", "Zoe"},
	Ignore:   "i",
	unexport: true,
}

func TestTagForm(t *testing.T) {
	a := assert.New(t)

	// Marshal
	data, err := Marshal(tagObjectData)
	a.NotError(err).
		Equal(string(data), formTagString)

	// Unmarshal
	obj := &tagObject{
		Ignore:   "i",
		unexport: true,
	}
	a.NotError(Unmarshal([]byte(formTagString), obj))
	a.Equal(obj, tagObjectData)
}
