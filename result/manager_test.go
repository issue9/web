// SPDX-License-Identifier: MIT

package result

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func init() {
	if err := message.SetString(language.Und, "lang", "und"); err != nil {
		panic(err)
	}

	if err := message.SetString(language.SimplifiedChinese, "lang", "hans"); err != nil {
		panic(err)
	}

	if err := message.SetString(language.TraditionalChinese, "lang", "hant"); err != nil {
		panic(err)
	}
}

func TestManager_NewResult(t *testing.T) {
	a := assert.New(t)
	mgr := NewManager(DefaultBuilder)
	mgr.AddMessage(400, 40000, "lang") // lang 有翻译

	// 能正常翻译错误信息
	rslt, ok := mgr.NewResult(message.NewPrinter(language.SimplifiedChinese), 40000).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "hans")

	// 采用 und
	rslt, ok = mgr.NewResult(message.NewPrinter(language.Und), 40000).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und")

	// 不存在的本地化信息，采用默认的 und
	rslt, ok = mgr.NewResult(message.NewPrinter(language.Afrikaans), 40000).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und")

	// 不存在
	a.Panic(func() { mgr.NewResult(message.NewPrinter(language.Afrikaans), 400) })
	a.Panic(func() { mgr.NewResult(message.NewPrinter(language.Afrikaans), 50000) })
}

func TestManager_NewResultWithFields(t *testing.T) {
	a := assert.New(t)
	mgr := NewManager(DefaultBuilder)
	mgr.AddMessage(400, 40000, "lang") // lang 有翻译
	fields := map[string][]string{"f1": {"v1", "v2"}}

	// 能正常翻译错误信息
	rslt, ok := mgr.NewResultWithFields(message.NewPrinter(language.SimplifiedChinese), 40000, fields).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "hans").
		Equal(rslt.Fields, []*fieldDetail{{Name: "f1", Message: []string{"v1", "v2"}}})

	// 采用 und
	rslt, ok = mgr.NewResultWithFields(message.NewPrinter(language.Und), 40000, fields).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und").
		Equal(rslt.Fields, []*fieldDetail{{Name: "f1", Message: []string{"v1", "v2"}}})
}

func TestManager_AddMessage(t *testing.T) {
	a := assert.New(t)
	mgr := NewManager(DefaultBuilder)

	a.NotPanic(func() {
		mgr.AddMessage(400, 1, "1")
		mgr.AddMessage(400, 100, "100")
	})

	msg, found := mgr.messages[1]
	a.True(found).
		Equal(msg.status, 400).
		Equal(msg.key, "1")

	msg, found = mgr.messages[401]
	a.False(found).Nil(msg)

	// 重复的 ID
	a.Panic(func() {
		mgr.AddMessage(400, 1, "40010")
	})
}

func TestManager_Messages(t *testing.T) {
	a := assert.New(t)
	mgr := NewManager(DefaultBuilder)
	a.NotNil(mgr)

	a.NotPanic(func() {
		mgr.AddMessage(400, 40010, "lang")
	})

	msg := mgr.Messages(message.NewPrinter(language.Und))
	a.Equal(msg[40010], "und")

	msg = mgr.Messages(message.NewPrinter(language.SimplifiedChinese))
	a.Equal(msg[40010], "hans")

	msg = mgr.Messages(message.NewPrinter(language.TraditionalChinese))
	a.Equal(msg[40010], "hant")

	msg = mgr.Messages(message.NewPrinter(language.English))
	a.Equal(msg[40010], "und")

	a.Panic(func() {
		mgr.Messages(nil)
	})
}
