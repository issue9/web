// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package plugintest

import (
	"sort"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web"
)

// 测试插件系统是否正常
func TestPlugins(t *testing.T) {
	a := assert.New(t)

	a.NotError(web.Classic("./testdata"))

	ms := web.Modules()
	a.Equal(3, len(ms)) // web-core 和两个插件

	sort.SliceStable(ms, func(i, j int) bool {
		return ms[i].Name < ms[j].Name
	})
	a.Equal(ms[0].Name, "plugin1")
	a.Equal(ms[1].Name, "plugin2")
}
