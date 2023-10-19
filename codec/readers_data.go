// 此文件由工具产生，请勿手动修改！

package codec

var (
	zstdInitData    = []byte{}
	gzipInitData    = []byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 0, 0, 0, 255, 255, 1, 0, 0, 255, 255, 0, 0, 0, 0, 0, 0, 0, 0}
	deflateInitData = []byte{0, 0, 0, 255, 255, 1, 0, 0, 255, 255}
	lzwInitData     = []byte{44}
	brotliInitData  = []byte{107, 0, 3}
)
