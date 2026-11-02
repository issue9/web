// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/logs/v7"
)

func TestLogsConfig_build(t *testing.T) {
	a := assert.New(t, false)

	e, f := buildFilter()
	conf := &logsConfig{}
	conf.build(f)
	a.False(f.HasError()).NotNil(conf.logs).Length(conf.cleanup, 0).
		Equal(conf.Levels, logs.AllLevels()).
		Empty(conf.Created).Empty(e)

	conf = &logsConfig{Levels: []logs.Level{logs.LevelWarn, logs.LevelError}, Created: logs.NanoLayout}
	e, f = buildFilter()
	conf.build(f)
	a.False(f.HasError()).NotNil(conf.logs).Length(conf.cleanup, 0).
		Equal(conf.Levels, []logs.Level{logs.LevelWarn, logs.LevelError}).
		Equal(conf.Created, logs.NanoLayout).
		False(conf.Location).Empty(e)
}

func TestLogsConfig_buildHandler(t *testing.T) {
	a := assert.New(t, false)

	// len(Handlers) == 0
	conf := &logsConfig{}
	_, f := buildFilter()
	h, c := conf.buildHandler(f)
	a.False(f.HasError()).Equal(h, logs.NewNopHandler()).Nil(c)

	// len(Handlers) == 1
	conf = &logsConfig{
		Levels: []logs.Level{logs.LevelInfo, logs.LevelError},
		Handlers: []*logHandlerConfig{
			{Type: "file", Args: []string{"./testdata", "2006%i.log", "1024", "text"}},
		},
	}
	_, f = buildFilter()
	h, c = conf.buildHandler(f)
	a.False(f.HasError()).NotNil(h).NotEqual(h, logs.NewNopHandler()).NotNil(c)

	// len(Handlers) > 1
	conf = &logsConfig{
		Levels: []logs.Level{logs.LevelInfo, logs.LevelError},
		Handlers: []*logHandlerConfig{
			{Type: "file", Args: []string{"./testdata", "2006%i.log", "1024", "text"}},
			{Type: "term", Args: []string{"stderr", "erro:red"}},
		},
	}
	_, f = buildFilter()
	h, c = conf.buildHandler(f)
	a.False(f.HasError()).NotNil(h).NotEqual(h, logs.NewNopHandler()).NotNil(c)
}

func TestNewTermHandler(t *testing.T) {
	a := assert.New(t, false)

	e, f := buildFilter()
	w, c := newTermLogsHandler(nil, f)
	a.True(f.HasError()).Nil(w).Nil(c)
	a.NotEmpty(e["args"])

	e, f = buildFilter()
	w, c = newTermLogsHandler([]string{"no-output", "red"}, f)
	a.True(f.HasError()).Nil(w).Nil(c)
	a.NotEmpty(e["args[0]"])

	e, f = buildFilter()
	w, c = newTermLogsHandler([]string{"stdout", "color-error"}, f)
	a.True(f.HasError()).Nil(w).Nil(c)
	a.NotEmpty(e["args[1]"])

	e, f = buildFilter()
	w, c = newTermLogsHandler([]string{"stdout", "erro:red"}, f)
	a.False(f.HasError()).NotNil(w).Nil(c).Empty(e)
}
