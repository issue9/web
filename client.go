// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/issue9/web/compress"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/mimetypes"
)

// Client 用于访问远程的客户端
type Client struct {
	mts    *mimetypes.Mimetypes[MarshalFunc, UnmarshalFunc]
	url    string
	client *http.Client

	marshal     func(any) ([]byte, error)
	marshalName string

	compresses *compress.Compresses
}

// NewClient 创建 Client 实例
//
// url 远程服务的地址基地址，url 不能以 / 结尾。比如 https://example.com:8080/s1；
// marshal 对输入数据的编码方式；
// mt 所有返回数据可用的解码方式；
func NewClient(url, marshalName string, marshal func(any) ([]byte, error), mt []*Mimetype, compresses []*Compress) (*Client, error) {
	if l := len(url); l > 0 && url[l-1] == '/' {
		url = url[:l-1]
	}

	mts := mimetypes.New[MarshalFunc, UnmarshalFunc](len(mt))
	for _, m := range mt {
		mts.Add(m.Type, m.Marshal, m.Unmarshal, m.ProblemType)
	}

	c := compress.NewCompresses(len(compresses))
	for i, e := range compresses {
		if err := e.sanitize(); err != nil {
			return nil, err.AddFieldParent("compresses[" + strconv.Itoa(i) + "]")
		}
		c.Add(e.Name, e.Compress, e.Types...)
	}

	return &Client{
		mts:    mts,
		url:    url,
		client: &http.Client{},

		marshal:     marshal,
		marshalName: marshalName,

		compresses: c,
	}, nil
}

// NewClient 采用 [Server] 的编码和压缩方式创建 Client 对象
//
// 参数可参考 [NewClient]。
func (srv *Server) NewClient(url, marshalName string, marshal func(any) ([]byte, error)) *Client {
	return &Client{
		mts:    srv.mimetypes,
		url:    url,
		client: &http.Client{},

		marshal:     marshal,
		marshalName: marshalName,

		compresses: srv.compresses,
	}
}

func (c *Client) Get(path string, resp any, problem *RFC7807) error {
	return c.NewRequest(http.MethodGet, path, nil, resp, problem)
}

func (c *Client) Delete(path string, resp any, problem *RFC7807) error {
	return c.NewRequest(http.MethodDelete, path, nil, resp, problem)
}

func (c *Client) Post(path string, req, resp any, problem *RFC7807) error {
	return c.NewRequest(http.MethodPost, path, req, resp, problem)
}

func (c *Client) Put(path string, req, resp any, problem *RFC7807) error {
	return c.NewRequest(http.MethodPut, path, req, resp, problem)
}

func (c *Client) Patch(path string, req, resp any, problem *RFC7807) error {
	return c.NewRequest(http.MethodPatch, path, req, resp, problem)
}

func (c *Client) NewRequest(method, path string, req, resp any, problem *RFC7807) (err error) {
	rsp, err := c.request(method, path, req)
	if err != nil {
		return err
	}

	var size int
	if h := rsp.Header.Get("Content-Length"); h != "" {
		if h == "0" {
			return nil
		}

		size, err = strconv.Atoi(h)
		if err != nil {
			return err
		}
	}
	if size == 0 {
		return nil
	}

	var reader io.Reader = rsp.Body
	encName := rsp.Header.Get(header.ContentEncoding)
	reader, err = c.compresses.ContentEncoding(encName, reader)
	if err != nil {
		return nil
	}

	var inputMimetype UnmarshalFunc
	var inputCharset encoding.Encoding
	if h := rsp.Header.Get(header.ContentType); h != "" {
		if inputMimetype, inputCharset, err = c.mts.ContentType(h); err != nil {
			return err
		}

		if inputMimetype == nil {
			return NewLocaleError("not found unmarshaler for the server content-type %s", h)
		}
	} else {
		return NewLocaleError("the server miss content-type header")
	}

	if !header.CharsetIsNop(inputCharset) {
		reader = transform.NewReader(reader, inputCharset.NewDecoder())
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil
	}
	defer rsp.Body.Close()

	if rsp.StatusCode >= 400 {
		return inputMimetype(data, problem)
	}
	return inputMimetype(data, resp)
}

func (c *Client) request(method, path string, req any) (resp *http.Response, err error) {
	var data []byte
	if req != nil {
		data, err = c.marshal(req)
		if err != nil {
			return nil, err
		}
	}

	url := c.url + path
	var r *http.Request
	if len(data) == 0 {
		r, err = http.NewRequest(method, url, nil)
	} else {
		r, err = http.NewRequest(method, url, bytes.NewBuffer(data))
	}
	if err != nil {
		return nil, err
	}
	r.Header.Set(header.ContentType, header.BuildContentType(c.marshalName, header.UTF8Name))
	r.Header.Set(header.Accept, c.mts.AcceptHeader())
	r.Header.Set(header.AcceptEncoding, c.compresses.AcceptEncodingHeader())

	return c.client.Do(r)
}
