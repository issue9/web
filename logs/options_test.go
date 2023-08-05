// SPDX-License-Identifier: MIT

package logs

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/config"
	"github.com/issue9/logs/v5"
)

func TestOptionsSanitize(t *testing.T) {
	a := assert.New(t, false)

	var o *Options
	o, err := optionsSanitize(o)
	a.NotError(err).NotNil(o)

	o = &Options{
		Levels:   []Level{logs.LevelDebug, logs.LevelError},
		StdLevel: 0,
	}
	o, err = optionsSanitize(o)
	a.NotError(err).NotNil(o)

	o = &Options{
		Levels:   []Level{logs.LevelDebug, 110},
		StdLevel: 0,
	}
	o, err = optionsSanitize(o)
	a.Equal(err.(*config.FieldError).Field, "Levels[1]").Nil(o)

	o = &Options{
		Levels:   []Level{logs.LevelDebug, logs.LevelError},
		StdLevel: 111,
	}
	o, err = optionsSanitize(o)
	a.Equal(err.(*config.FieldError).Field, "StdLevel").Nil(o)
}