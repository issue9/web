// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package install

import "github.com/issue9/term/colors"

const (
	infoForeground = colors.Green
	infoBackground = colors.Default

	warnForeground = colors.Yellow
	warnBackground = colors.Default

	errorForeground = colors.Red
	errorBackground = colors.Default
)

// Logger 输出日志
type Logger struct {
}

func (l *Logger) Print(f, b colors.Color, v ...interface{}) error {
	_, err := colors.Print(f, b, v...)
	return err
}

func (l *Logger) Println(f, b colors.Color, v ...interface{}) error {
	_, err := colors.Println(f, b, v...)
	return err
}

func (l *Logger) Printf(f, b colors.Color, format string, v ...interface{}) error {
	_, err := colors.Printf(f, b, format, v...)
	return err
}

func (l *Logger) Info(v ...interface{}) {
	l.Print(infoForeground, infoBackground, v...)
}

func (l *Logger) Infoln(v ...interface{}) {
	l.Println(infoForeground, infoBackground, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Printf(infoForeground, infoBackground, format, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.Print(warnForeground, warnBackground, v...)
}

func (l *Logger) Warnln(v ...interface{}) {
	l.Println(warnForeground, warnBackground, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Printf(warnForeground, warnBackground, format, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.Print(errorForeground, errorBackground, v...)
}

func (l *Logger) Errorln(v ...interface{}) {
	l.Println(errorForeground, errorBackground, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Printf(errorForeground, errorBackground, format, v...)
}
