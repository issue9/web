// SPDX-License-Identifier: MIT

package form

import (
	"encoding"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web"
)

var (
	sexValue                          = Sex(1)
	_        encoding.TextMarshaler   = sexValue
	_        encoding.TextUnmarshaler = &sexValue
)

type (
	Sex int8

	TagObject struct {
		Name     string
		Age      int      `form:"age"`
		Friend   []string `form:"friend"`
		Ignore   string   `form:"-"`
		Sex      Sex      `form:"sex"`
		unexport bool
	}

	anonymousObject struct {
		*TagObject
		Address string `form:"address"`
		F       func()
	}

	nestObject struct {
		Tag  *TagObject `form:"tags"`
		Maps map[string]*anonymousObject
	}
)

const tagObjectString = "Name=Ava&age=10&friend=Jess&friend=Sarah&friend=Zoe&sex=female"

var tagObjectData = &TagObject{
	Name:     "Ava",
	Age:      10,
	Friend:   []string{"Jess", "Sarah", "Zoe"},
	Sex:      1,
	Ignore:   "i",
	unexport: true,
}

const anonymousString = "Name=Ava&address=1&age=10&friend=Jess&friend=Sarah&friend=Zoe&sex=male"

var anonymousData = &anonymousObject{
	TagObject: &TagObject{
		Name:   "Ava",
		Age:    10,
		Friend: []string{"Jess", "Sarah", "Zoe"},
	},
	Address: "1",
}

const nestString = "Maps.1.Name=Ava&Maps.1.address=1&Maps.1.age=10&Maps.1.friend=Jess&Maps.1.friend=Sarah&Maps.1.sex=male&Maps.2.Name=Ava&Maps.2.address=1&Maps.2.age=10&Maps.2.friend=Jess&Maps.2.friend=Zoe&Maps.2.sex=female&tags.Name=Ava&tags.age=10&tags.friend=Jess&tags.friend=Sarah&tags.friend=Zoe&tags.sex=male"

var nestData = &nestObject{
	Tag: &TagObject{
		Name:   "Ava",
		Age:    10,
		Friend: []string{"Jess", "Sarah", "Zoe"},
	},
	Maps: map[string]*anonymousObject{
		"1": {
			TagObject: &TagObject{
				Name:   "Ava",
				Age:    10,
				Friend: []string{"Jess", "Sarah"},
			},
			Address: "1",
		},
		"2": {
			TagObject: &TagObject{
				Name:   "Ava",
				Age:    10,
				Friend: []string{"Jess", "Zoe"},
				Sex:    1,
			},
			Address: "1",
		},
	},
}

func (s Sex) MarshalText() ([]byte, error) {
	switch s {
	case 0:
		return []byte("male"), nil
	case 1:
		return []byte("female"), nil
	default:
		return nil, web.ErrUnsupportedSerialization()
	}
}

func (s *Sex) UnmarshalText(v []byte) error {
	switch string(v) {
	case "male":
		*s = 0
	case "female":
		*s = 1
	default:
		return web.ErrUnsupportedSerialization()
	}
	return nil
}

func TestMarshalWithFormTag(t *testing.T) {
	a := assert.New(t, false)

	// Marshal
	data, err := BuildMarshal(nil)(tagObjectData)
	a.NotError(err).Equal(string(data), tagObjectString)

	// Unmarshal
	obj := &TagObject{
		Ignore:   "i",
		unexport: true,
	}
	a.NotError(Unmarshal([]byte(tagObjectString), obj))
	a.Equal(obj, tagObjectData)

	// anonymous marshal
	data, err = BuildMarshal(nil)(anonymousData)
	a.NotError(err).
		Equal(string(data), anonymousString)

	// anonymous unmarshal
	anoobj := &anonymousObject{}
	a.NotError(Unmarshal([]byte(anonymousString), anoobj))
	a.Equal(anoobj, anonymousData)

	// nest marshal
	data, err = BuildMarshal(nil)(nestData)
	a.NotError(err).
		Equal(string(data), nestString)

	// nest unmarshal
	nestObj := &nestObject{}
	a.NotError(Unmarshal([]byte(nestString), nestObj))
	a.Equal(nestObj, nestData)
}
