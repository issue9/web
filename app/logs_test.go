// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/logs/v7"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

func TestLogsConfig_build(t *testing.T) {
	a := assert.New(t, false)

	conf := &logsConfig{}
	err := conf.build()
	a.NotError(err).NotNil(conf.logs).Length(conf.cleanup, 0).
		Equal(conf.logs.Levels, server.AllLevels()).
		Empty(conf.logs.Created)

	conf = &logsConfig{Levels: []logs.Level{logs.LevelWarn, logs.LevelError}, Created: logs.NanoLayout}
	err = conf.build()
	a.NotError(err).NotNil(conf.logs).Length(conf.cleanup, 0).
		Equal(conf.logs.Levels, []logs.Level{logs.LevelWarn, logs.LevelError}).
		Equal(conf.logs.Created, logs.NanoLayout).
		False(conf.logs.Location)
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
