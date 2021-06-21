// SPDX-License-Identifier: MIT

package text

import (
	"encoding"
	"testing"

	"github.com/issue9/assert"
)

var (
	_ encoding.TextMarshaler   = &TestObject{}
	_ encoding.TextUnmarshaler = &TestObject{}
)

func TestTextObject(t *testing.T) {
	a := assert.New(t)

	obj := &TestObject{Name: "name", Age: 1}
	data, err := obj.MarshalText()
	a.NotError(err).
		NotNil(data).
		Equal(string(data), "name,1")

	a.NotError(obj.UnmarshalText([]byte("unmarshal,22")))
	a.Equal(obj.Name, "unmarshal").Equal(obj.Age, 22)

	a.Error(obj.UnmarshalText([]byte("unmarshal,22,33")))

	a.Error(obj.UnmarshalText([]byte("22,unmarshal")))
}
