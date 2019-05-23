// 由工具自动生成，无需手动修改！

package version

// Version 版本号
//
// 版本号规则遵循 https://semver.org/lang/zh-CN/
const Version = "0.28.3"

// buildDate 编译日期，可以由编译器指定
var buildDate string

var fullVersion = Version

func init() {
	if buildDate != "" {
		fullVersion = Version + "+" + buildDate
	}
}

// FullVersion 完整的版本号，可能包括了编译日期。
func FullVersion() string {
	return fullVersion
}
