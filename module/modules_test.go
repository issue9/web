// SPDX-License-Identifier: MIT

package module

import (
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/config"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/app"
	"github.com/issue9/web/internal/resulttest"
	"github.com/issue9/web/internal/webconfig"
)

func job(time.Time) error {
	println("job")
	return nil
}

func newModules(a *assert.Assertion) *Modules {
	var configUnmarshals = map[string]config.UnmarshalFunc{
		".yaml": yaml.Unmarshal,
		".yml":  yaml.Unmarshal,
	}

	mgr, err := config.NewManager("./testdata")
	a.NotError(err).NotNil(mgr)
	for k, v := range configUnmarshals {
		a.NotError(mgr.AddUnmarshal(v, k))
	}

	webconf := &webconfig.WebConfig{}
	a.NotError(mgr.LoadFile("web.yaml", webconf))

	app := app.New(webconf, getResult)
	a.NotNil(app)

	ms, err := NewModules(app, "")
	a.NotError(err).NotNil(ms)
	return ms
}

func getResult(status, code int, message string) app.Result {
	return resulttest.New(status, code, message)
}
