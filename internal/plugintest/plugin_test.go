// SPDX-License-Identifier: MIT

// +build linux darwin freebsd

package plugintest

import (
	"sort"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/config"
)

// 测试插件系统是否正常
func TestPlugins(t *testing.T) {
	a := assert.New(t)

	w, err := config.Classic("./testdata/logs.xml", "./testdata/web.yaml")
	a.NotError(err).NotNil(w)

	ms := w.Modules()
	a.Equal(2, len(ms))

	sort.SliceStable(ms, func(i, j int) bool {
		return ms[i].ID() < ms[j].ID()
	})
	a.Equal(ms[0].ID(), "plugin1")
	a.Equal(ms[1].ID(), "plugin2")
}
