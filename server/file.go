// SPDX-License-Identifier: MIT

package server

import (
	"io/fs"
	"log"
	"strconv"

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/locale"
)

func (conf *configOf[T]) sanitizeFileSerializers() *web.FieldError {
	for i, name := range conf.FileSerializers {
		s, found := filesFactory.get(name)
		if !found {
			return web.NewFieldError("["+strconv.Itoa(i)+"]", web.NewLocaleError("not found serialization function for %s", name))
		}
		conf.config.Serializers = append(conf.config.Serializers, s)
	}
	return nil
}

func loadConfigOf[T any](configDir, name string) (*configOf[T], error) {
	c, err := config.BuildDir(buildSerializerFromFactory(), configDir)
	if err != nil {
		return nil, err
	}

	conf := &configOf[T]{config: &Config{Dir: configDir}}
	if err := c.Load(name, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func buildSerializerFromFactory() config.Serializer {
	s := make(config.Serializer, len(filesFactory.items))
	for ext, item := range filesFactory.items {
		s.Add(item.Marshal, item.Unmarshal, ext)
	}
	return s
}

// NewPrinter 根据参数构建一个本地化的打印对象
//
// 语言由 [localeutil.DetectUserLanguageTag] 决定。
// fsys 指定了加载本地化文件的文件系统，glob 则指定了加载的文件匹配规则；
// 对于文件的序列化方式则是根据后缀名从由 [RegisterFileSerializer] 注册的项中查找。
func NewPrinter(glob string, fsys ...fs.FS) (*localeutil.Printer, error) {
	tag, err := localeutil.DetectUserLanguageTag()
	if err != nil {
		log.Println(err) // 输出错误，但是不中断执行
	}

	b := catalog.NewBuilder(catalog.Fallback(tag))
	if err := locale.Load(buildSerializerFromFactory(), b, glob, fsys...); err != nil {
		return nil, err
	}

	return locale.NewPrinter(tag, b), nil
}
