// SPDX-License-Identifier: MIT

//go:generate go run ./make_data.go

// Package compressor 提供了所有支持的压缩算法
package compressor

import "io"

// Compressor 压缩算法的接口
type Compressor interface {
	// Name 算法的名称
	Name() string

	// NewDecoder 将 r 包装成为当前压缩算法的解码器
	NewDecoder(r io.Reader) (io.ReadCloser, error)

	// NewEncoder 将 w 包装成当前压缩算法的编码器
	NewEncoder(w io.Writer) (io.WriteCloser, error)
}
