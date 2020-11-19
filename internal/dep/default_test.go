// SPDX-License-Identifier: MIT

package dep

func newMod(id string, f func() error, dep ...string) *Default {
	d := NewDefaultModule(id, id+" description", dep...)
	d.AddInit(f, id)
	return d
}
