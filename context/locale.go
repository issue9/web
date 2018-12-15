// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"io"

	"golang.org/x/text/message"
)

// Fprint 相当于 ctx.LocalePrinter.Fprint
func (ctx *Context) Fprint(w io.Writer, v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Fprint(w, v...)
}

// Fprintf 相当于 ctx.LocalePrinter.Fprintf
func (ctx *Context) Fprintf(w io.Writer, key message.Reference, v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Fprintf(w, key, v...)
}

// Fprintln 相当于 ctx.LocalePrinter.Fprintln
func (ctx *Context) Fprintln(w io.Writer, v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Fprintln(w, v...)
}

// Print 相当于 ctx.LocalePrinter.Print
func (ctx *Context) Print(v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Print(v...)
}

// Printf 相当于 ctx.LocalePrinter.Printf
func (ctx *Context) Printf(key message.Reference, v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Printf(key, v...)
}

// Println 相当于 ctx.LocalePrinter.Println
func (ctx *Context) Println(v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Println(v...)
}

// Sprint 相当于 ctx.LocalePrinter.Sprint
func (ctx *Context) Sprint(v ...interface{}) string {
	return ctx.LocalePrinter.Sprint(v...)
}

// Sprintf 相当于 ctx.LocalePrinter.Sprintf
func (ctx *Context) Sprintf(key message.Reference, v ...interface{}) string {
	return ctx.LocalePrinter.Sprintf(key, v...)
}

// Sprintln 相当于 ctx.LocalePrinter.Sprintln
func (ctx *Context) Sprintln(v ...interface{}) string {
	return ctx.LocalePrinter.Sprintln(v...)
}