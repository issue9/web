// SPDX-License-Identifier: MIT

package mimetype

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/issue9/qheader"
)

type marshaler struct {
	f    MarshalFunc
	name string
}

type unmarshaler struct {
	f    UnmarshalFunc
	name string
}

func mimetypeExists(name string) error {
	return fmt.Errorf("该名称 %s 的 mimetype 已经存在", name)
}

// Mimetypes 管理 mimetype 的增删改查
type Mimetypes struct {
	marshals   []*marshaler
	unmarshals []*unmarshaler
}

// NewMimetypes 返回 *Mimetypes 实例
func NewMimetypes() *Mimetypes {
	return &Mimetypes{
		marshals:   make([]*marshaler, 0, 10),
		unmarshals: make([]*unmarshaler, 0, 10),
	}
}

// Unmarshal 查找指定名称的 UnmarshalFunc
func (srv *Mimetypes) Unmarshal(name string) (UnmarshalFunc, error) {
	var unmarshal *unmarshaler
	for _, mt := range srv.unmarshals {
		if mt.name == name {
			unmarshal = mt
			break
		}
	}
	if unmarshal == nil {
		return nil, fmt.Errorf("未找到 %s 类型的解码函数", name)
	}

	return unmarshal.f, nil
}

// Marshal 从 header 解析出当前请求所需要的解 mimetype 名称和对应的解码函数
//
// */* 或是空值 表示匹配任意内容，一般会选择第一个元素作匹配；
// xx/* 表示匹配以 xx/ 开头的任意元素，一般会选择 xx/* 开头的第一个元素；
// xx/ 表示完全匹配以 xx/ 的内容
// 如果传递的内容如下：
//  application/json;q=0.9,*/*;q=1
// 则因为 */* 的 q 值比较高，而返回 */* 匹配的内容
//
// 在不完全匹配的情况下，返回值的名称依然是具体名称。
//  text/*;q=0.9
// 返回的名称可能是：
//  text/plain
func (srv *Mimetypes) Marshal(header string) (string, MarshalFunc, error) {
	if header == "" {
		if mm := srv.findMarshal("*/*"); mm != nil {
			return mm.name, mm.f, nil
		}
		return "", nil, errors.New("请求中未指定 accept 报头，且服务端也未指定匹配 */* 的解码函数")
	}

	accepts := qheader.Parse(header, "*/*")
	for _, accept := range accepts {
		if mm := srv.findMarshal(accept.Value); mm != nil {
			return mm.name, mm.f, nil
		}
	}

	return "", nil, errors.New("未找到符合客户端要求的解码函数")
}

// AddMarshals 添加多个编码函数
func (srv *Mimetypes) AddMarshals(ms map[string]MarshalFunc) error {
	for k, v := range ms {
		if err := srv.AddMarshal(k, v); err != nil {
			return err
		}
	}

	return nil
}

// AddMarshal 添加编码函数
//
// mf 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP
// 另作处理，比如下载，上传等内容。
func (srv *Mimetypes) AddMarshal(name string, mf MarshalFunc) error {
	if strings.HasSuffix(name, "/*") || name == "*" {
		panic("name 不是一个有效的 mimetype 名称格式")
	}

	for _, mt := range srv.marshals {
		if mt.name == name {
			return mimetypeExists(name)
		}
	}

	srv.marshals = append(srv.marshals, &marshaler{
		f:    mf,
		name: name,
	})

	sort.SliceStable(srv.marshals, func(i, j int) bool {
		if srv.marshals[i].name == DefaultMimetype {
			return true
		}

		if srv.marshals[j].name == DefaultMimetype {
			return false
		}

		return srv.marshals[i].name < srv.marshals[j].name
	})

	return nil
}

// AddUnmarshals 添加多个编码函数
func (srv *Mimetypes) AddUnmarshals(ms map[string]UnmarshalFunc) error {
	for k, v := range ms {
		if err := srv.AddUnmarshal(k, v); err != nil {
			return err
		}
	}

	return nil
}

// AddUnmarshal 添加编码函数
//
// mm 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP
// 另作处理，比如下载，上传等内容。
func (srv *Mimetypes) AddUnmarshal(name string, mm UnmarshalFunc) error {
	if strings.IndexByte(name, '*') >= 0 {
		panic("name 不是一个有效的 mimetype 名称格式")
	}

	for _, mt := range srv.unmarshals {
		if mt.name == name {
			return mimetypeExists(name)
		}
	}

	srv.unmarshals = append(srv.unmarshals, &unmarshaler{
		f:    mm,
		name: name,
	})

	sort.SliceStable(srv.unmarshals, func(i, j int) bool {
		if srv.unmarshals[i].name == DefaultMimetype {
			return true
		}

		if srv.unmarshals[j].name == DefaultMimetype {
			return false
		}

		return srv.unmarshals[i].name < srv.unmarshals[j].name
	})

	return nil
}

func (srv *Mimetypes) findMarshal(name string) *marshaler {
	switch {
	case len(srv.marshals) == 0:
		return nil
	case name == "" || name == "*/*":
		return srv.marshals[0] // 由 len(marshals) == 0 确保最少有一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		for _, mt := range srv.marshals {
			if strings.HasPrefix(mt.name, prefix) {
				return mt
			}
		}
	default:
		for _, mt := range srv.marshals {
			if mt.name == name {
				return mt
			}
		}
	}
	return nil
}
