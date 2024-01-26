// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"testing"

	"github.com/issue9/assert/v3"
)

func RegisterRegisterFileSerializer(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		RegisterFileSerializer("new", json.Marshal, json.Unmarshal, ".json")
	}, "扩展名 .json 已经注册到 json")

	RegisterFileSerializer("new", json.Marshal, json.Unmarshal, ".js")
	v, f := filesFactory.get("new")
	a.True(f).NotNil(v).Equal(v.Exts, []string{".js"})
}
