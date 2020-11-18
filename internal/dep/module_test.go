// SPDX-License-Identifier: MIT

package dep

type mod struct {
	id     string
	deps   []string
	inited bool
	f      func()
}

var _ Module = &mod{}

func newMod(id string, f func(), dep ...string) *mod {
	return &mod{id: id, f: f, deps: dep}
}

func (m *mod) ID() string {
	return m.id
}

func (m *mod) Deps() []string {
	return m.deps
}

func (m *mod) Inited() bool {
	return m.inited
}

func (m *mod) Init() error {
	m.f()
	m.inited = true
	return nil
}
