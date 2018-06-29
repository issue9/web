// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package form

import (
	"net/url"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/encoding"
)

var (
	_ encoding.MarshalFunc   = Marshal
	_ encoding.UnmarshalFunc = Unmarshal
)

func TestMarshal(t *testing.T) {
	a := assert.New(t)

	v := url.Values{}
	v.Set("name", "Ava")
	v.Add("friend", "Jess")
	v.Add("friend", "Sarah")
	v.Add("friend", "Zoe")

	data, err := Marshal(v)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "friend=Jess&friend=Sarah&friend=Zoe&name=Ava")
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t)

	v := url.Values{}
	a.NotError(Unmarshal([]byte("friend=Jess&friend=Sarah&friend=Zoe&name=Ava"), v))
	a.Equal(v.Get("name"), "Ava")
	a.Equal(v.Get("friend"), "Jess")
	a.Equal(v["friend"], []string{"Jess", "Sarah", "Zoe"})
}
