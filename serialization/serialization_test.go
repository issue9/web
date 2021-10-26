// SPDX-License-Identifier: MIT

package serialization

import (
	"encoding/json"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/localeutil"
)

func TestSerialization(t *testing.T) {
	a := assert.New(t)
	s := New(5)
	a.NotNil(s)

	// 不能添加同名的多次
	a.NotError(s.Add(nil, nil, "n1", "n2"))
	a.Equal(2, s.Len())
	a.Equal(s.Add(nil, nil, "n1"), localeutil.Error("has serialization function %s", "n1"))
	a.Equal(2, s.Len())

	// set
	s.Set("n1", json.Marshal, json.Unmarshal)
	a.Equal(2, s.Len())
	s.Set("n3", json.Marshal, json.Unmarshal)
	a.Equal(3, s.Len())

	// search
	n, m, u := s.Search("n1")
	a.Equal(n, "n1").NotNil(m).NotNil(u)

	// 删除
	s.Delete("n1")
	s.Delete("not-exists")
	a.Equal(2, s.Len())
	n, m, u = s.Search("n1")
	a.Empty(n).Nil(m).Nil(u)
}
