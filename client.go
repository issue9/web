// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"io"
	"net/http"

	"github.com/issue9/mux/v10/header"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/qheader"
	"github.com/issue9/web/internal/status"
	"github.com/issue9/web/selector"
)

// Client 用于访问远程的客户端
//
// NOTE: 远程如果不是 [Server] 实现的服务，可能无法正确处理返回对象。
type Client struct {
	client   *http.Client
	codec    *Codec
	selector selector.Selector

	marshal     func(any) ([]byte, error)
	marshalName string

	requestIDKey string
	requestIDGen func() string
}

// NewClient 创建 [Client] 实例
//
// client 如果为空，表示采用 &http.Client{} 作为默认值；
// marshalName 和 marshal 表示编码的名称和方法；
// requestIDKey 表示 x-request-id 的报头名称，如果为空则表示不需要；
// requestIDGen 表示生成 x-request-id 值的方法；
func NewClient(
	client *http.Client,
	codec *Codec,
	s selector.Selector,
	marshalName string,
	marshal func(any) ([]byte, error),
	requestIDKey string,
	requestIDGen func() string,
) *Client {
	if client == nil {
		client = &http.Client{}
	}

	if requestIDKey != "" && requestIDGen == nil {
		panic("当前 requestIDKey 不为空时 requestIDGen 也不能为空")
	}

	return &Client{
		client:   client,
		codec:    codec,
		selector: s,

		marshalName: marshalName,
		marshal:     marshal,

		requestIDKey: requestIDKey,
		requestIDGen: requestIDGen,
	}
}

func (c *Client) Get[R any, E any](path string) (*R, error) {
	return c.Do[R, E](http.MethodGet, path, nil)
}

func (c *Client) Delete[R any, E any](path string) (*R, error) {
	return c.Do[R, E](http.MethodDelete, path, nil)
}

func (c *Client) Post[R any, E any](path string, body any) (*R, error) {
	return c.Do[R, E](http.MethodPost, path, body)
}

func (c *Client) Put[R any, E any](path string, body any) (*R, error) {
	return c.Do[R, E](http.MethodPut, path, body)
}

func (c *Client) Patch[R any, E any](path string, body any) (*R, error) {
	return c.Do[R, E](http.MethodPatch, path, body)
}

// Do 开始新的请求
//
// body 为提交的对象，最终是由初始化参数的 marshal 进行编码；
func (c *Client) Do[R any, E any](method, path string, body any) (*R, error) {
	r, err := c.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	rsp, err := c.Client().Do(r)
	if err != nil {
		return nil, err
	}

	return c.ParseResponse[R, E](rsp)
}

// ParseResponse 从 [http.Response] 解析并获取返回对象
//
// R 表示正常解析之后返回的类型，该值不能为指针类型；
// E 表示出错时返回的 Problem.Extensions 的类型，大部分情况该值为 any 即可；
func (c *Client) ParseResponse[R any, E any](rsp *http.Response) (r *R, err error) {
	if rsp.ContentLength == 0 { // 204 可能为空
		return nil, nil
	}

	var reader io.Reader = rsp.Body
	encName := rsp.Header.Get(header.ContentEncoding)
	reader, err = c.codec.contentEncoding(encName, reader)
	if err != nil {
		return nil, err
	}

	var inputMimetype UnmarshalFunc
	var inputCharset encoding.Encoding
	if h := rsp.Header.Get(header.ContentType); h != "" {
		if inputMimetype, inputCharset, err = c.codec.contentType(h); err != nil {
			return nil, err
		}

		if inputMimetype == nil {
			return nil, NewLocaleError("not found unmarshaler for the server content-type %s", h)
		}
	} else {
		return nil, NewLocaleError("the server miss content-type header")
	}

	if !qheader.CharsetIsNop(inputCharset) {
		reader = transform.NewReader(reader, inputCharset.NewDecoder())
	}

	if status.IsProblemStatus(rsp.StatusCode) {
		var e E
		p := &Problem{Extensions: e}
		if err := inputMimetype(reader, p); err != nil {
			return nil, err
		}
		return nil, p
	}

	var resp R
	if err = inputMimetype(reader, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// NewRequest 生成 [http.Request]
//
// body 为需要提交的对象；
func (c *Client) NewRequest(method, path string, body any) (resp *http.Request, err error) {
	var data []byte
	if body != nil {
		data, err = c.marshal(body)
		if err != nil {
			return nil, err
		}
	}

	if path, err = c.URL(path); err != nil {
		return nil, err
	}

	var r *http.Request
	if len(data) == 0 {
		r, err = http.NewRequest(method, path, nil)
	} else {
		r, err = http.NewRequest(method, path, bytes.NewBuffer(data))
	}
	if err != nil {
		return nil, err
	}
	r.Header.Set(header.ContentType, qheader.BuildContentType(c.marshalName, header.UTF8))
	r.Header.Set(header.Accept, c.codec.clientAcceptHeader)
	r.Header.Set(header.AcceptEncoding, c.codec.acceptEncodingHeader)
	if c.requestIDKey != "" {
		r.Header.Set(c.requestIDKey, c.requestIDGen())
	}

	return r, nil
}

// URL 生成一条访问地址
func (c *Client) URL(path string) (string, error) {
	u, err := c.selector.Next()
	if err != nil {
		return "", err
	}
	return u + path, nil
}

func (c *Client) Client() *http.Client { return c.client }
