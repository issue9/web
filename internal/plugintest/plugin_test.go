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

	srv, err := config.Classic("./testdata/logs.xml", "./testdata/web.yaml")
	a.NotError(err).NotNil(srv)

	ms := srv.Modules()
	a.Equal(2, len(ms))

	sort.SliceStable(ms, func(i, j int) bool {
		return ms[i].ID() < ms[j].ID()
	})
	a.Equal(ms[0].ID(), "plugin1")
	a.Equal(ms[1].ID(), "plugin2")

	// 手动加载插件
	a.NotError(srv.LoadPlugin("./testdata/plugin3.so"))
	ms = srv.Modules()
	a.Equal(3, len(ms)).Equal(ms[2].ID(), "plugin3")

	// 加载已经加载的插件
	a.Error(srv.LoadPlugins("./testdata/plugin*.so"))
	ms = srv.Modules()
	a.Equal(3, len(ms))
}
