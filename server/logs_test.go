// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/logs/v7"

	"github.com/issue9/web"
)

func TestLogsConfig_build(t *testing.T) {
	a := assert.New(t, false)

	conf := &logsConfig{}
	err := conf.build()
	a.NotError(err).NotNil(conf.logs).Length(conf.cleanup, 0).
		Equal(conf.logs.Levels, AllLevels()).
		Empty(conf.logs.Created)

	conf = &logsConfig{Levels: []logs.Level{logs.LevelWarn, logs.LevelError}, Created: logs.NanoLayout}
	err = conf.build()
	a.NotError(err).NotNil(conf.logs).Length(conf.cleanup, 0).
		Equal(conf.logs.Levels, []logs.Level{logs.LevelWarn, logs.LevelError}).
		Equal(conf.logs.Created, logs.NanoLayout).
		False(conf.logs.Location)
}

func TestLogsConfig_buildHandler(t *testing.T) {
	a := assert.New(t, false)

	// len(Handlers) == 0
	conf := &logsConfig{}
	h, c, err := conf.buildHandler()
	a.NotError(err).Equal(h, NewNopHandler()).Nil(c)

	// len(Handlers) == 1
	conf = &logsConfig{
		Levels: []logs.Level{logs.LevelInfo, logs.LevelError},
		Handlers: []*logHandlerConfig{
			{Type: "file", Args: []string{"./testdata", "2006%i.log", "1024", "text"}},
		},
	}
	h, c, err = conf.buildHandler()
	a.NotError(err).NotNil(h).NotEqual(h, NewNopHandler()).NotNil(c)

	// len(Handlers) > 1
	conf = &logsConfig{
		Levels: []logs.Level{logs.LevelInfo, logs.LevelError},
		Handlers: []*logHandlerConfig{
			{Type: "file", Args: []string{"./testdata", "2006%i.log", "1024", "text"}},
			{Type: "term", Args: []string{"stderr", "erro:red"}},
		},
	}
	h, c, err = conf.buildHandler()
	a.NotError(err).NotNil(h).NotEqual(h, NewNopHandler()).NotNil(c)
}

func TestNewTermHandler(t *testing.T) {
	a := assert.New(t, false)

	w, c, err := newTermLogsHandler(nil)
	a.Error(err).Nil(w).Nil(c)
	ce, ok := err.(*web.FieldError)
	a.True(ok).Equal(ce.Field, "args")

	w, c, err = newTermLogsHandler([]string{"no-output", "red"})
	a.Error(err).Nil(w).Nil(c)
	ce, ok = err.(*web.FieldError)
	a.True(ok).Equal(ce.Field, "args[0]")

	w, c, err = newTermLogsHandler([]string{"stdout", "color-error"})
	a.Error(err).Nil(w).Nil(c)
	ce, ok = err.(*web.FieldError)
	a.True(ok).Equal(ce.Field, "args[1]")

	w, c, err = newTermLogsHandler([]string{"stdout", "erro:red"})
	a.NotError(err).NotNil(w).Nil(c)
}
