// SPDX-License-Identifier: MIT

package main

import (
	"embed"
	"fmt"

	"github.com/issue9/localeutil"
	"github.com/issue9/localeutil/message"
	xmessage "golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"
)

//go:embed locales/*.yaml
var locales embed.FS

func newPrinter() (*localeutil.Printer, error) {
	tag, err := localeutil.DetectUserLanguageTag()
	if err != nil {
		fmt.Println(err)
	}

	m := &message.Messages{}
	if err = m.LoadFSGlob(locales, "locales/*.yaml", yaml.Unmarshal); err != nil {
		return nil, err
	}

	b := catalog.NewBuilder()
	if err = m.Catalog(b); err != nil {
		return nil, err
	}

	return xmessage.NewPrinter(tag, xmessage.Catalog(b)), nil
}
