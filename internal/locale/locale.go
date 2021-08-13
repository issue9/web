// SPDX-License-Identifier: MIT

package locale

import (
	"encoding/json"
	"io/fs"

	"golang.org/x/text/message/catalog"
)

// Load 从文件加载本地化文件
func LoadFS(b *catalog.Builder, f fs.FS, glob string) error {
	matches, err := fs.Glob(f, glob)
	if err != nil {
		return err
	}

	for _, p := range matches {
		data, err := fs.ReadFile(f, p)
		if err != nil {
			return err
		}

		if err := Load(b, data); err != nil {
			return err
		}
	}

	return nil
}

// Load 加 data 的内容至 b
func Load(b *catalog.Builder, data []byte) (err error) {
	l := &Messages{}
	if err := json.Unmarshal(data, l); err != nil {
		return err
	}

	for _, msg := range l.Messages {
		switch {
		case msg.Message.Vars != nil:
			vars := msg.Message.Vars
			msgs := make([]catalog.Message, 0, len(vars))
			for _, v := range vars {
				msgs = append(msgs, v.message())
			}
			msgs = append(msgs, catalog.String(msg.Message.Msg))
			err = b.Set(l.Language, msg.Key, msgs...)
		case msg.Message.Select != nil:
			s := msg.Message.Select
			err = b.Set(l.Language, msg.Key, s.message())
		case msg.Message.Msg != "":
			err = b.SetString(l.Language, msg.Key, msg.Message.Msg)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
