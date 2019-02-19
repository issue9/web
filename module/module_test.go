// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/internal/webconfig"
)

func TestTag(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)
	m := newModule(ms, TypeModule, "user1", "user1 desc")
	a.NotNil(m).Equal(m.Type, TypeModule)

	v := m.NewTag("0.1.0")
	a.NotNil(v).NotNil(m.tags["0.1.0"])
	a.Equal(v.Type, TypeTag).Equal(v.Name, "0.1.0")
	v.AddInit(nil, "title1")
	a.Equal(v.inits[0].title, "title1")

	vv := m.NewTag("0.1.0")
	a.Equal(vv, v).Equal(vv.Name, "0.1.0")

	v2 := m.NewTag("0.2.0")
	a.NotEqual(v2, v)

	// 子标签，不能再添加子标
	a.Panic(func() {
		vv.NewTag("0.3.0")
	})
}
