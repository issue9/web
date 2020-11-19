// SPDX-License-Identifier: MIT

package dep

func newMod(id string, f func() error, dep ...string) *Default {
	d := NewDefaultModule(id, dep...)
	d.AddInit(id, f)
	return d
}
