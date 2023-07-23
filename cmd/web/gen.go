// SPDX-License-Identifier: MIT

package main

//go:generate web locale -l=en-US -f=yaml ./
//go:generate web update-locale -src=./locales/en-US.yaml -dest=./locales/zh-CN.yaml
