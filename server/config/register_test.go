// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestRegisterFileSerializer(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		RegisterFileSerializer("new", json.Marshal, json.Unmarshal, ".json")
	}, "扩展名 .json 已经注册到 json")

	RegisterFileSerializer("new", json.Marshal, json.Unmarshal, ".js")
	v, f := fileSerializerFactory.get("new")
	a.True(f).NotNil(v).Equal(v.exts, []string{".js"})
}
