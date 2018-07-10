// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"plugin"
)

func loadPlugin(path string) (*Module, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	m := &Module{}

	// Name
	symbol, err := p.Lookup("ModuleName")
	if err != nil {
		return nil, err
	}
	m.Name = *symbol.(*string)

	// Description
	symbol, err = p.Lookup("ModuleDescription")
	if err != nil {
		return nil, err
	}
	m.Description = *symbol.(*string)

	// Deps
	symbol, err = p.Lookup("ModuleDeps")
	if err != nil {
		return nil, err
	}
	m.Deps = *symbol.(*[]string)

	// inits
	symbol, err = p.Lookup("ModuleInits")
	if err != nil {
		return nil, err
	}
	m.inits = *symbol.(*[]func() error)

	// TODO installs, routes
	return m, nil
}
