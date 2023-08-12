// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/issue9/localeutil"
	"github.com/issue9/localeutil/message/serialize"
	wl "github.com/issue9/web/locales"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/cmd/web/locales"
)

func newPrinter() (*localeutil.Printer, error) {
	tag, err := localeutil.DetectUserLanguageTag()
	if err != nil {
		fmt.Println(err)
	}

	ls, err := serialize.LoadFSGlob(locales.Locales, "*.yaml", yaml.Unmarshal)
	if err != nil {
		return nil, err
	}

	webLocales, err := serialize.LoadFSGlob(wl.Locales, "*.yaml", yaml.Unmarshal)
	if err != nil {
		return nil, err
	}

	ls = append(ls, webLocales...)

	b := catalog.NewBuilder()
	for _, l := range ls {
		if err = l.Catalog(b); err != nil {
			return nil, err
		}
	}

	return message.NewPrinter(tag, message.Catalog(b)), nil
}
