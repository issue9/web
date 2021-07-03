// SPDX-License-Identifier: MIT

package content

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/text/encoding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/transform"
)

// 需要作比较，所以得是经过 http.CanonicalHeaderKey 处理的标准名称。
var (
	contentTypeKey     = http.CanonicalHeaderKey("Content-Type")
	contentLanguageKey = http.CanonicalHeaderKey("Content-Language")
)

// Context 单次请求生成的上下文数据
type Context struct {
	Response http.ResponseWriter
	Request  *http.Request

	// 指定输出时所使用的媒体类型，以及名称
	OutputMimetype     MarshalFunc
	OutputMimetypeName string

	// 输出到客户端的字符集
	//
	// 若值为 encoding.Nop 或是空，表示为 utf-8
	OutputCharset     encoding.Encoding
	OutputCharsetName string

	// 客户端内容所使用的媒体类型
	InputMimetype UnmarshalFunc

	// 客户端内容所使用的字符集
	//
	// 若值为 encoding.Nop 或是空，表示为 utf-8
	InputCharset encoding.Encoding

	// 输出语言的相关设置项
	OutputTag     language.Tag
	LocalePrinter *message.Printer

	// 保存着从 http.Request.Body 中获取的内容。
	//
	// body 用于缓存从 http.Request.Body 中读取的内容；
	// read 表示是否需要从 http.Request.Body 读取内容。
	body []byte
	read bool
}

// NewContext 从用户请求中构建一个 Context 实例
//
// 如果不合规则，会以指定的状码退出。
func (c *Content) NewContext(w http.ResponseWriter, r *http.Request) (*Context, int) {
	header := r.Header.Get("Accept")
	outputMimetypeName, marshal, found := c.Marshal(header)
	if !found {
		return nil, http.StatusNotAcceptable
	}

	header = r.Header.Get("Accept-Charset")
	outputCharsetName, outputCharset := AcceptCharset(header)
	if outputCharsetName == "" {
		return nil, http.StatusNotAcceptable
	}

	header = r.Header.Get(contentTypeKey)
	inputMimetype, inputCharset, err := c.ConentType(header)
	if err != nil {
		return nil, http.StatusUnsupportedMediaType
	}

	tag := AcceptLanguage(c.CatalogBuilder(), r.Header.Get("Accept-Language"))

	ctx := &Context{
		Response:           w,
		Request:            r,
		OutputMimetype:     marshal,
		OutputMimetypeName: outputMimetypeName,
		OutputCharset:      outputCharset,
		OutputCharsetName:  outputCharsetName,
		OutputTag:          tag,
		LocalePrinter:      c.NewLocalePrinter(tag),
		InputMimetype:      inputMimetype,
		InputCharset:       inputCharset,
	}

	return ctx, 0
}

// Body 获取用户提交的内容
//
// 相对于 ctx.Request.Body，此函数可多次读取。不存在 body 时，返回 nil
func (ctx *Context) Body() (body []byte, err error) {
	if ctx.read {
		return ctx.body, nil
	}

	if ctx.body, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
		return nil, err
	}
	ctx.read = true

	if CharsetIsNop(ctx.InputCharset) {
		return ctx.body, nil
	}

	d := ctx.InputCharset.NewDecoder()
	reader := transform.NewReader(bytes.NewReader(ctx.body), d)
	ctx.body, err = ioutil.ReadAll(reader)
	return ctx.body, err
}

// Unmarshal 将提交的内容转换成 v 对象
func (ctx *Context) Unmarshal(v interface{}) error {
	body, err := ctx.Body()
	if err != nil {
		return err
	}

	if ctx.InputMimetype != nil {
		return ctx.InputMimetype(body, v)
	}
	return nil
}

// Marshal 将 v 解码并发送给客户端
//
// NOTE: 若 v 是一个 nil 值，则不会向客户端输出任何内容；
// 若是需要正常输出一个 nil 类型到客户端（比如JSON 中的 null），
// 可以传递一个 *struct{} 值，或是自定义实现相应的解码函数。
//
// NOTE: 如果需要指定一个特定的 Content-Type 和 Content-Language，
// 可以在 headers 中指定，否则使用当前的编码和语言名称。
func (ctx *Context) Marshal(status int, v interface{}, headers map[string]string) error {
	header := ctx.Response.Header()
	var contentTypeFound, contentLanguageFound bool
	for k, v := range headers {
		k = http.CanonicalHeaderKey(k)

		contentTypeFound = contentTypeFound || k == contentTypeKey
		contentLanguageFound = contentLanguageFound || k == contentLanguageKey
		header.Set(k, v)
	}

	if !contentTypeFound {
		ct := BuildContentType(ctx.OutputMimetypeName, ctx.OutputCharsetName)
		header.Set(contentTypeKey, ct)
	}

	if !contentLanguageFound && ctx.OutputTag != language.Und {
		header.Set(contentLanguageKey, ctx.OutputTag.String())
	}

	if v == nil {
		ctx.Response.WriteHeader(status)
		return nil
	}

	data, err := ctx.OutputMimetype(v)
	if err != nil {
		return err
	}

	// 注意 WriteHeader 调用顺序。
	// https://github.com/golang/go/issues/17083
	//
	// NOTE: 此处由原来的 errorhandler.WriteHeader 改为 ctx.Response.WriteHeader
	// 即 Marshal 函数也接受 errorhandler 的捕获，不作特殊处理。
	ctx.Response.WriteHeader(status)

	if CharsetIsNop(ctx.OutputCharset) {
		_, err = ctx.Response.Write(data)
		return err
	}

	w := transform.NewWriter(ctx.Response, ctx.OutputCharset.NewEncoder())
	if _, err = w.Write(data); err != nil {
		if err2 := w.Close(); err2 != nil {
			return fmt.Errorf("在处理错误 %w 时再次抛出错误 %s", err, err2.Error())
		}
		return err
	}
	return w.Close()
}
