// SPDX-License-Identifier: MIT

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

type nestObject struct {
	Tag  *TagObject `form:"tags"`
	Maps map[string]*anonymousObject
}

const formTagString = "Name=Ava&age=10&friend=Jess&friend=Sarah&friend=Zoe"

var tagObjectData = &TagObject{
	Name:     "Ava",
	Age:      10,
	Friend:   []string{"Jess", "Sarah", "Zoe"},
	Ignore:   "i",
	unexport: true,
}

const anonymousString = "Name=Ava&address=1&age=10&friend=Jess&friend=Sarah&friend=Zoe"

var anonymousData = &anonymousObject{
	TagObject: &TagObject{
		Name:   "Ava",
		Age:    10,
		Friend: []string{"Jess", "Sarah", "Zoe"},
	},
	Address: "1",
}

const nestString = "Maps.1.Name=Ava&Maps.1.address=1&Maps.1.age=10&Maps.1.friend=Jess&Maps.1.friend=Sarah&Maps.2.Name=Ava&Maps.2.address=1&Maps.2.age=10&Maps.2.friend=Jess&Maps.2.friend=Zoe&tags.Name=Ava&tags.age=10&tags.friend=Jess&tags.friend=Sarah&tags.friend=Zoe"

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
			},
			Address: "1",
		},
	},
}

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

	// anonymous unmarshal
	anoobj := &anonymousObject{}
	a.NotError(Unmarshal([]byte(anonymousString), anoobj))
	a.Equal(anoobj, anonymousData)

	// nest marshal
	data, err = Marshal(nestData)
	a.NotError(err).
		Equal(string(data), nestString)

	// nest unmarshal
	nestObj := &nestObject{}
	a.NotError(Unmarshal([]byte(nestString), nestObj))
	a.Equal(nestObj, nestData)
}
