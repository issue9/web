// SPDX-License-Identifier: MIT

package locale

import (
	"encoding/json"
	"io/fs"

	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
)

type locale struct {
	XMLName  struct{}     `json:"-" xml:"locale" yaml:"-"`
	ID       language.Tag `json:"id" xml:"id,attr" yaml:"id"`
	Messages []*message   `json:"messages" xml:"messages" yaml:"messages"`
}

type message struct {
	Key   string            `json:"key" xml:"key" yaml:"key"`
	Value []catalog.Message `json:"value,omitempty" xml:"value,omitempty" yaml:"value,omitempty"`
}

// Load 从文件加载本地化文件
func Load(b *catalog.Builder, f fs.FS, glob string) error {
	matches, err := fs.Glob(f, glob)
	if err != nil {
		return err
	}

	for _, p := range matches {
		data, err := fs.ReadFile(f, p)
		if err != nil {
			return err
		}

		l := &locale{}
		if err := json.Unmarshal(data, l); err != nil {
			return err
		}

		for _, msg := range l.Messages {
			if err := b.Set(l.ID, msg.Key, msg.Value...); err != nil {
				return err
			}
		}
	}

	return nil
}
