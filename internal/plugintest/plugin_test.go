// SPDX-License-Identifier: MIT

//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package plugintest

import (
	"sort"
	"testing"
	"time"

	"github.com/issue9/assert"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

// 测试插件系统是否正常
func TestPlugins(t *testing.T) {
	a := assert.New(t)

	srv, err := web.NewServer("app", "0.1.0", &web.Options{
		Plugins: "./testdata/plugin_*.so",
	})
	a.NotError(err).NotNil(srv)

	go func() {
		srv.Serve("default", true)
	}()
	time.Sleep(500 * time.Millisecond)
	defer srv.Close(0)

	ms := srv.Modules(message.NewPrinter(language.SimplifiedChinese))
	a.Equal(2, len(ms))

	sort.SliceStable(ms, func(i, j int) bool {
		return ms[i].ID < ms[j].ID
	})
	a.Equal(ms[0].ID, "plugin1")
	a.Equal(ms[1].ID, "plugin2")
}
