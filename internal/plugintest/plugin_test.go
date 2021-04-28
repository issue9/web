// SPDX-License-Identifier: MIT

// +build linux darwin freebsd

package plugintest

import (
	"sort"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v2"

	"github.com/issue9/web"
)

// 测试插件系统是否正常
func TestPlugins(t *testing.T) {
	a := assert.New(t)

	srv, err := web.NewServer("app", "0.1.0", logs.New(), &web.Options{})
	a.NotError(err).NotNil(srv)

	a.NotError(srv.LoadPlugins("./testdata/plugin_*.so"))

	go func() {
		srv.Serve()
	}()
	time.Sleep(500 * time.Millisecond)
	defer srv.Close(0)

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
