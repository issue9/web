// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"encoding/json"
	"testing"

	"github.com/issue9/assert"
)

func TestMarshal(t *testing.T) {
	a := assert.New(t)

	a.Nil(Marshal("not exists"))
	a.NotNil(Marshal(DefaultMimeType))

	// 添加已存在的
	a.Equal(AddMarshal(DefaultMimeType, json.Marshal), ErrExists)

	a.NotError(AddMarshal("json", json.Marshal))
	a.NotNil(Marshal("json"))
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t)

	a.Nil(Unmarshal("not exists"))
	a.NotNil(Unmarshal(DefaultMimeType))

	// 添加已存在的
	a.Equal(AddUnmarshal(DefaultMimeType, json.Unmarshal), ErrExists)

	a.NotError(AddUnmarshal("json", json.Unmarshal))
	a.NotNil(Unmarshal("json"))
}
