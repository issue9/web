// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web"
	"github.com/issue9/web/logs"
)

func TestLogsConfig_build(t *testing.T) {
	a := assert.New(t, false)

	conf := &logsConfig{}
	o, c, err := conf.build()
	a.NotError(err).NotNil(o).Length(c, 0).
		Equal(o.Levels, logs.AllLevels()).
		Empty(o.Created)

	conf = &logsConfig{Levels: []logs.Level{logs.Warn, logs.Error}, Created: logs.NanoLayout}
	o, c, err = conf.build()
	a.NotError(err).NotNil(o).Length(c, 0).
		Equal(o.Levels, []logs.Level{logs.Warn, logs.Error}).
		Equal(o.Created, logs.NanoLayout).
		False(o.Location)
}

func TestLogsConfig_output(t *testing.T) {
	a := assert.New(t, false)

	conf := &logsConfig{
		Created: "2006",
		Handlers: []*logHandlerConfig{
			{
				Type: "file",
				Args: []string{"./testdata", "1504-%i.log", "1024"},
			},
			{
				Type: "term",
				Args: []string{"stdout"},
			},
			{
				Type: "term",
				Args: []string{"stdout", "erro:red", "warn:yellow"},
			},
		},
	}
	o, c, err := conf.build()
	a.NotError(err).NotNil(o).Length(c, 1) // 文件有 cleanup 返回
	l, err1 := logs.New(nil, o)
	a.NotError(err1).NotNil(l)
	l.ERROR().Print("test")
	a.NotError(c[0]())
}

func TestNewTermHandler(t *testing.T) {
	a := assert.New(t, false)

	w, c, err := newTermLogsHandler(nil)
	a.Error(err).Nil(w).Nil(c)
	ce, ok := err.(*web.FieldError)
	a.True(ok).Equal(ce.Field, "Args")

	w, c, err = newTermLogsHandler([]string{"no-output", "red"})
	a.Error(err).Nil(w).Nil(c)
	ce, ok = err.(*web.FieldError)
	a.True(ok).Equal(ce.Field, "Args[0]")

	w, c, err = newTermLogsHandler([]string{"stdout", "color-error"})
	a.Error(err).Nil(w).Nil(c)
	ce, ok = err.(*web.FieldError)
	a.True(ok).Equal(ce.Field, "Args[1]")

	w, c, err = newTermLogsHandler([]string{"stdout", "erro:red"})
	a.NotError(err).NotNil(w).Nil(c)
}
