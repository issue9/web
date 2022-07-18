// SPDX-License-Identifier: MIT

package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/serializer"
)

type Serializer struct {
	s *serializer.Serializer
}

func NewSerializer(s *serializer.Serializer) *Serializer { return &Serializer{s: s} }

func (f *Serializer) Serializer() *serializer.Serializer { return f.s }

func (f *Serializer) Save(p string, v any) error {
	m, _ := f.SearchByExt(p)
	if m == nil {
		return localeutil.Error("not found serialization function for %s", p)
	}

	data, err := m(v)
	if err != nil {
		return err
	}

	return os.WriteFile(p, data, fs.ModePerm)
}

func (f *Serializer) Load(fsys fs.FS, name string, v any) error {
	_, u := f.SearchByExt(name)
	if u == nil {
		return localeutil.Error("not found serialization function for %s", name)
	}

	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return err
	}

	return u(data, v)
}

func (f *Serializer) SearchByExt(filename string) (serializer.MarshalFunc, serializer.UnmarshalFunc) {
	ext := filepath.Ext(filename)
	_, m, u := f.Serializer().Search(ext)
	return m, u
}
