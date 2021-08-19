// SPDX-License-Identifier: MIT

// +build linux darwin freebsd

package plugintest

import (
	"sort"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/web"
)

// 测试插件系统是否正常
func TestPlugins(t *testing.T) {
	a := assert.New(t)

	srv, err := web.NewServer("app", "0.1.0", &web.Options{
		Plugins: "./testdata/plugin_*.so",
	})
	a.NotError(err).NotNil(srv)
	a.NotError(srv.InitModules("default"))

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
}
