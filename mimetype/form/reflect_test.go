// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package form

import (
	"testing"

	"github.com/issue9/assert"
)

type TagObject struct {
	Name     string
	Age      int      `form:"age"`
	Friend   []string `form:"friend"`
	Ignore   string   `form:"-"`
	unexport bool
}

type anonymousObject struct {
	*TagObject
	Address string `form:"address"`
	F       func()
}

var (
	formTagString = "Name=Ava&age=10&friend=Jess&friend=Sarah&friend=Zoe"

	tagObjectData = &TagObject{
		Name:     "Ava",
		Age:      10,
		Friend:   []string{"Jess", "Sarah", "Zoe"},
		Ignore:   "i",
		unexport: true,
	}
)

var (
	anonymousString = "Name=Ava&address=1&age=10&friend=Jess&friend=Sarah&friend=Zoe"
	anonymousData   = &anonymousObject{
		TagObject: &TagObject{
			Name:   "Ava",
			Age:    10,
			Friend: []string{"Jess", "Sarah", "Zoe"},
		},
		Address: "1",
	}
)

func TestTagForm(t *testing.T) {
	a := assert.New(t)

	// Marshal
	data, err := Marshal(tagObjectData)
	a.NotError(err).
		Equal(string(data), formTagString)

	// Unmarshal
	obj := &TagObject{
		Ignore:   "i",
		unexport: true,
	}
	a.NotError(Unmarshal([]byte(formTagString), obj))
	a.Equal(obj, tagObjectData)

	// anonymous marhsal
	data, err = Marshal(anonymousData)
	a.NotError(err).
		Equal(string(data), anonymousString)

	anoobj := &anonymousObject{}
	a.NotError(Unmarshal([]byte(anonymousString), anoobj))
	a.Equal(anoobj, anonymousData)
}
