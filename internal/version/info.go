// 由工具自动生成，不能修改！

package version

// Version 版本号
//
// 版本号规则遵循 https://semver.org/lang/zh-CN/
const Version = "0.31.0"

// 编译日期，可以由编译器指定
var buildDate string

// 最后一次提交的 hash 值
var commitHash string

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

// CommitHash 最后一次提示我的 hash 值
func FullVersion() string {
	return fullVersion
}
