// SPDX-License-Identifier: MIT

package dep

var _ Module = &Default{}

func newMod(id string, f func() error, dep ...string) *Default {
	d := NewDefaultModule(id, id+" description", dep...)
	d.AddInit(f, id)
	return d
}
