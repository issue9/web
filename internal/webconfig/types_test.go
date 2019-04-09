// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package webconfig

import (
	"encoding/xml"
	"testing"

	"github.com/issue9/assert"
)

var (
	_ xml.Marshaler   = &pairs{}
	_ xml.Unmarshaler = &pairs{}
)

type testStruct struct {
	Pairs pairs `xml:"pairs"`
}

func TestPairs(t *testing.T) {
	a := assert.New(t)

	m := &testStruct{
		Pairs: pairs{ // 多个字段，注意 map 顺序问题
			"key1": "val1",
		},
	}

	bs, err := xml.MarshalIndent(m, "", "  ")
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<testStruct>
  <pairs>
    <key name="key1">val1</key>
  </pairs>
</testStruct>`)

	rm := &testStruct{}
	xml.Unmarshal(bs, rm)
	a.Equal(rm, m)

	// 空值
	m = &testStruct{
		Pairs: pairs{},
	}

	bs, err = xml.MarshalIndent(m, "", "  ")
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<testStruct></testStruct>`)
}
