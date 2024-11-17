// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"fmt"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/text/message"
)

type Callback struct {
	Ref      *Ref
	Callback map[string]*PathItem
}

type callbackRenderer = orderedmap.OrderedMap[string, *renderer[pathItemRenderer]]

func (c *Callback) build(p *message.Printer, d *Document) *renderer[callbackRenderer] {
	if c.Ref != nil {
		return newRenderer[callbackRenderer](c.Ref.build(p, "callbacks"), nil)
	}
	return newRenderer(nil, c.buildRenderer(p, d))
}

func (c *Callback) buildRenderer(p *message.Printer, d *Document) *callbackRenderer {
	return writeMap2OrderedMap(c.Callback, nil, func(in *PathItem) *renderer[pathItemRenderer] { return in.build(p, d, nil) })
}

func (resp *Callback) addComponents(c *components) {
	if resp.Ref != nil {
		if _, found := c.callbacks[resp.Ref.Ref]; !found {
			c.callbacks[resp.Ref.Ref] = resp
		}
	}

	for _, item := range resp.Callback {
		item.addComponents(c)
	}
}

// AddWebhook 添加 Webhook 的定义
func (d *Document) AddWebhook(name, method string, o *Operation) {
	if d.webHooks == nil {
		d.webHooks = make(map[string]*PathItem, 5)
	}

	hook, found := d.webHooks[name]
	if !found {
		hook = &PathItem{}
		d.webHooks[name] = hook
	}

	if hook.Operations == nil {
		hook.Operations = make(map[string]*Operation, 3)
	} else if _, found := hook.Operations[method]; found {
		panic(fmt.Sprintf("已经存在 %s:%s 的 webhook", name, method))
	}
	hook.Operations[method] = o
}
