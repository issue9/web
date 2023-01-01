// SPDX-License-Identifier: MIT

package form

import (
	"net/url"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/server"
)

var (
	_ server.MarshalFunc   = Marshal
	_ server.UnmarshalFunc = Unmarshal

	_ Marshaler   = objectData
	_ Unmarshaler = objectData
)

var formString = "friend=Jess&friend=Sarah&friend=Zoe&name=Ava"

var objectData = &object{
	Name:   "Ava",
	Friend: []string{"Jess", "Sarah", "Zoe"},
}

type object struct {
	Name   string
	Friend []string
}

func (obj *object) MarshalForm() ([]byte, error) {
	vals := url.Values{}

	vals.Set("name", obj.Name)
	for _, v := range obj.Friend {
		vals.Add("friend", v)
	}

	return []byte(vals.Encode()), nil
}

func (obj *object) UnmarshalForm(data []byte) error {
	vals, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	obj.Name = vals.Get("name")
	obj.Friend = vals["friend"]

	return nil
}

func TestMarshal(t *testing.T) {
	a := assert.New(t, false)

	formObject := url.Values{}
	data, err := Marshal(nil, formObject)
	a.NotError(err)
	a.NotNil(data). // 非 nil
			Empty(data) // 但长度为 0

	formObject.Set("name", "Ava")
	formObject.Add("friend", "Jess")
	formObject.Add("friend", "Sarah")
	formObject.Add("friend", "Zoe")
	data, err = Marshal(nil, formObject)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), formString)

	// 非 url.Values 类型
	data, err = Marshal(nil, &struct{}{})
	a.NotError(err).Empty(data)

	// Marshaler 类型
	data, err = Marshal(nil, objectData)
	a.NotError(err).
		Equal(string(data), formString)
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t, false)

	v := url.Values{}
	a.NotError(Unmarshal(nil, v))
	a.Equal(len(v), 0)

	v = url.Values{}
	a.NotError(Unmarshal([]byte{}, v))
	a.Equal(len(v), 0)

	v = url.Values{}
	a.Error(Unmarshal([]byte("%"), v))

	a.NotError(Unmarshal([]byte(formString), &struct{}{}))

	v = url.Values{}
	a.NotError(Unmarshal([]byte(formString), v))
	a.Equal(len(v), 2)
	a.Equal(v.Get("name"), "Ava")
	a.Equal(v.Get("friend"), "Jess")
	a.Equal(v["friend"], []string{"Jess", "Sarah", "Zoe"})

	// Unmarshaler 类型
	obj := &object{}
	a.NotError(Unmarshal([]byte(formString), obj))
	a.Equal(obj, objectData)
}
