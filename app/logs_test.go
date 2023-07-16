// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/config"

	"github.com/issue9/web/logs"
)

func TestLogsConfig_build(t *testing.T) {
	a := assert.New(t, false)

	conf := &logsConfig{}
	o, c, err := conf.build()
	a.NotError(err).NotNil(o).Length(c, 0).
		Equal(o.Levels, logs.AllLevels()).
		False(o.Created)

	conf = &logsConfig{Levels: []logs.Level{logs.Warn, logs.Error}, Created: true}
	o, c, err = conf.build()
	a.NotError(err).NotNil(o).Length(c, 0).
		Equal(o.Levels, []logs.Level{logs.Warn, logs.Error}).
		True(o.Created).
		False(o.Caller)
}

func TestLogsConfig_output(t *testing.T) {
	a := assert.New(t, false)

	conf := &logsConfig{
		Handlers: []*logHandlerConfig{
			{
				Type: "file",
				Args: []string{"2006", "./testdata", "1504-%i.log", "1024"},
			},
			{
				Type: "term",
				Args: []string{"2006", "stdout"},
			},
			{
				Type: "term",
				Args: []string{"2006", "stdout", "erro:red", "warn:yellow"},
			},
		},
	}
	o, c, err := conf.build()
	a.NotError(err).NotNil(o).Length(c, 1) // 文件有 cleanup 返回
	l, err1 := logs.New(o)
	a.NotError(err1).NotNil(l)
	l.ERROR().Print("test")
	a.NotError(c[0]())
}

func TestNewTermHandler(t *testing.T) {
	a := assert.New(t, false)

	w, c, err := newTermLogsHandler(nil)
	a.Error(err).Nil(w).Nil(c)
	ce, ok := err.(*config.FieldError)
	a.True(ok).Equal(ce.Field, "Args")

	w, c, err = newTermLogsHandler([]string{"2006", "no-output", "no-color"})
	a.Error(err).Nil(w).Nil(c)
	ce, ok = err.(*config.FieldError)
	a.True(ok).Equal(ce.Field, "Args[1]")

	w, c, err = newTermLogsHandler([]string{"2006", "stdout", "color-error"})
	a.Error(err).Nil(w).Nil(c)
	ce, ok = err.(*config.FieldError)
	a.True(ok).Equal(ce.Field, "Args[2]")

	w, c, err = newTermLogsHandler([]string{"2006", "stdout", "erro:red"})
	a.NotError(err).NotNil(w).Nil(c)
}
