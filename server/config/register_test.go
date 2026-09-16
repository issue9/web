// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json/v2"
	"testing"

	"github.com/issue9/assert/v5"
)

func jm(a any) ([]byte, error) { return json.Marshal(a) }

func ju(b []byte, a any) error { return json.Unmarshal(b, a) }

func TestRegisterFileSerializer(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		RegisterFileSerializer("new", jm, ju, ".json")
	}, "扩展名 .json 已经注册到 json")

	RegisterFileSerializer("new", jm, ju, ".js")
	v, f := fileSerializerFactory.get("new")
	a.True(f).NotNil(v).Equal(v.exts, []string{".js"})
}
