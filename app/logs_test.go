// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/internal/errs"
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
		Writers: []*logWriterConfig{
			{
				Type: "file",
				Args: []string{"2006", "./testdata", "1504-%i.log", "1024"},
			},
			{
				Type: "term",
				Args: []string{"2006", "red", "stdout"},
			},
		},
	}
	o, c, err := conf.build()
	a.NotError(err).NotNil(o).Length(c, 1) // 文件有 cleanup 返回
	logs.New(o, nil).ERROR().Print("test")
	a.NotError(c[0]())
}

func TestNewTermWriter(t *testing.T) {
	a := assert.New(t, false)

	w, c, err := newTermLogsWriter(nil)
	a.Error(err).Nil(w).Nil(c)
	ce, ok := err.(*errs.ConfigError)
	a.True(ok).Equal(ce.Field, "Args")

	w, c, err = newTermLogsWriter([]string{"2006", "no-color", "no-output"})
	a.Error(err).Nil(w).Nil(c)
	ce, ok = err.(*errs.ConfigError)
	a.True(ok).Equal(ce.Field, "Args[1]")

	w, c, err = newTermLogsWriter([]string{"2006", "default", "no-output"})
	a.Error(err).Nil(w).Nil(c)
	ce, ok = err.(*errs.ConfigError)
	a.True(ok).Equal(ce.Field, "Args[2]")

	w, c, err = newTermLogsWriter([]string{"2006", "default", "stdout"})
	a.NotError(err).NotNil(w).Nil(c)
}
