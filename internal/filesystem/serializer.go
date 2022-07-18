// SPDX-License-Identifier: MIT

package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/issue9/localeutil"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/serializer"
)

type Serializer struct {
	s *serializer.Serializer
}

func NewSerializer(s *serializer.Serializer) *Serializer { return &Serializer{s: s} }

func (f *Serializer) Serializer() *serializer.Serializer { return f.s }

func (f *Serializer) Save(p string, v any) error {
	m, _ := f.searchByExt(p)
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
	_, u := f.searchByExt(name)
	if u == nil {
		return localeutil.Error("not found serialization function for %s", name)
	}

	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return err
	}

	return u(data, v)
}

func (f *Serializer) searchByExt(filename string) (serializer.MarshalFunc, serializer.UnmarshalFunc) {
	ext := filepath.Ext(filename)
	_, m, u := f.Serializer().Search(ext)
	return m, u
}

func LoadLocaleFiles(f *Serializer, b *catalog.Builder, fsys fs.FS, glob string) error {
	matches, err := fs.Glob(fsys, glob)
	if err != nil {
		return err
	}

	for _, m := range matches {
		_, u := f.searchByExt(m)
		if u == nil {
			return localeutil.Error("not found serialization function for %s", m)
		}

		if err := localeutil.LoadMessageFromFS(b, fsys, m, u); err != nil {
			return err
		}
	}

	return nil
}
