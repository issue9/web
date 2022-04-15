// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/logs/v4"
)

func TestLogsConfig_build(t *testing.T) {
	a := assert.New(t, false)

	conf := &logsConfig{}
	l, c, err := conf.build()
	a.NotError(err).NotNil(l).Length(c, 0).
		True(l.IsEnable(logs.LevelError)).
		True(l.IsEnable(logs.LevelInfo)).
		False(l.HasCreated()).
		Equal(6, len(conf.Levels))

	conf = &logsConfig{Levels: []logs.Level{logs.LevelWarn, logs.LevelError}, Created: true}
	l, c, err = conf.build()
	a.NotError(err).NotNil(l).Length(c, 0).
		True(l.IsEnable(logs.LevelError)).
		False(l.IsEnable(logs.LevelInfo)).
		True(l.HasCreated()).
		False(l.HasCaller()).
		Equal(2, len(conf.Levels))
}

func TestLogsConfig_output(t *testing.T) {
	a := assert.New(t, false)

	conf := &logsConfig{
		Writers: []*logWritterConfig{
			{
				Type: "file",
				Args: []string{"2006", "1504-%i.log", "./testdata", "1024"},
			},
			{
				Type: "term",
				Args: []string{"2006", "red", "stdout"},
			},
		},
	}
	l, c, err := conf.build()
	a.NotError(err).NotNil(l).Length(c, 1) // 文件有 cleanup 返回
	l.ERROR().Print("test")
	a.NotError(c[0]())
}

func TestNewTermWriter(t *testing.T) {
	a := assert.New(t, false)

	w, c, err := newTermLogsWriter(nil)
	a.Error(err).Nil(w).Nil(c)
	ce, ok := err.(*ConfigError)
	a.True(ok).Equal(ce.Field, "Args")

	w, c, err = newTermLogsWriter([]string{"2006", "no-color", "no-output"})
	a.Error(err).Nil(w).Nil(c)
	ce, ok = err.(*ConfigError)
	a.True(ok).Equal(ce.Field, "Args[1]")

	w, c, err = newTermLogsWriter([]string{"2006", "default", "no-output"})
	a.Error(err).Nil(w).Nil(c)
	ce, ok = err.(*ConfigError)
	a.True(ok).Equal(ce.Field, "Args[2]")

	w, c, err = newTermLogsWriter([]string{"2006", "default", "stdout"})
	a.NotError(err).NotNil(w).Nil(c)
}
