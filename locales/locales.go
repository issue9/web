// SPDX-License-Identifier: MIT

// Package locales 为 web 包提供了本地化的内容
package locales

import (
	"embed"

	"github.com/issue9/localeutil"
)

//go:embed *.yml
var Locales embed.FS

// 一些多处用到的翻译项
const (
	ShouldGreatThanZero = localeutil.StringPhrase("should great than 0")
	InvalidValue        = localeutil.StringPhrase("invalid value")
	CanNotBeEmpty       = localeutil.StringPhrase("can not be empty")
	DuplicateValue      = localeutil.StringPhrase("duplicate value")
)

// problem 的文档说明
const (
	Problem400       = localeutil.StringPhrase("problem.400")
	Problem400Detail = localeutil.StringPhrase("problem.400.detail")

	Problem401       = localeutil.StringPhrase("problem.401")
	Problem401Detail = localeutil.StringPhrase("problem.401.detail")

	Problem402       = localeutil.StringPhrase("problem.402")
	Problem402Detail = localeutil.StringPhrase("problem.402.detail")

	Problem403       = localeutil.StringPhrase("problem.403")
	Problem403Detail = localeutil.StringPhrase("problem.403.detail")

	Problem404       = localeutil.StringPhrase("problem.404")
	Problem404Detail = localeutil.StringPhrase("problem.404.detail")

	Problem405       = localeutil.StringPhrase("problem.405")
	Problem405Detail = localeutil.StringPhrase("problem.405.detail")

	Problem406       = localeutil.StringPhrase("problem.406")
	Problem406Detail = localeutil.StringPhrase("problem.406.detail")

	Problem407       = localeutil.StringPhrase("problem.407")
	Problem407Detail = localeutil.StringPhrase("problem.407.detail")

	Problem408       = localeutil.StringPhrase("problem.408")
	Problem408Detail = localeutil.StringPhrase("problem.408.detail")

	Problem409       = localeutil.StringPhrase("problem.409")
	Problem409Detail = localeutil.StringPhrase("problem.409.detail")

	Problem410       = localeutil.StringPhrase("problem.410")
	Problem410Detail = localeutil.StringPhrase("problem.410.detail")

	Problem411       = localeutil.StringPhrase("problem.411")
	Problem411Detail = localeutil.StringPhrase("problem.411.detail")

	Problem412       = localeutil.StringPhrase("problem.412")
	Problem412Detail = localeutil.StringPhrase("problem.412.detail")

	Problem413       = localeutil.StringPhrase("problem.413")
	Problem413Detail = localeutil.StringPhrase("problem.413.detail")

	Problem414       = localeutil.StringPhrase("problem.414")
	Problem414Detail = localeutil.StringPhrase("problem.414.detail")

	Problem415       = localeutil.StringPhrase("problem.415")
	Problem415Detail = localeutil.StringPhrase("problem.415.detail")

	Problem416       = localeutil.StringPhrase("problem.416")
	Problem416Detail = localeutil.StringPhrase("problem.416.detail")

	Problem417       = localeutil.StringPhrase("problem.417")
	Problem417Detail = localeutil.StringPhrase("problem.417.detail")

	Problem418       = localeutil.StringPhrase("problem.418")
	Problem418Detail = localeutil.StringPhrase("problem.418.detail")

	Problem421       = localeutil.StringPhrase("problem.421")
	Problem421Detail = localeutil.StringPhrase("problem.421.detail")

	Problem422       = localeutil.StringPhrase("problem.422")
	Problem422Detail = localeutil.StringPhrase("problem.422.detail")

	Problem423       = localeutil.StringPhrase("problem.423")
	Problem423Detail = localeutil.StringPhrase("problem.423.detail")

	Problem424       = localeutil.StringPhrase("problem.424")
	Problem424Detail = localeutil.StringPhrase("problem.424.detail")

	Problem425       = localeutil.StringPhrase("problem.425")
	Problem425Detail = localeutil.StringPhrase("problem.425.detail")

	Problem426       = localeutil.StringPhrase("problem.426")
	Problem426Detail = localeutil.StringPhrase("problem.426.detail")

	Problem428       = localeutil.StringPhrase("problem.428")
	Problem428Detail = localeutil.StringPhrase("problem.428.detail")

	Problem429       = localeutil.StringPhrase("problem.429")
	Problem429Detail = localeutil.StringPhrase("problem.429.detail")

	Problem431       = localeutil.StringPhrase("problem.431")
	Problem431Detail = localeutil.StringPhrase("problem.431.detail")

	Problem451       = localeutil.StringPhrase("problem.451")
	Problem451Detail = localeutil.StringPhrase("problem.451.detail")

	Problem500       = localeutil.StringPhrase("problem.500")
	Problem500Detail = localeutil.StringPhrase("problem.500.detail")

	Problem501       = localeutil.StringPhrase("problem.501")
	Problem501Detail = localeutil.StringPhrase("problem.501.detail")

	Problem502       = localeutil.StringPhrase("problem.502")
	Problem502Detail = localeutil.StringPhrase("problem.502.detail")

	Problem503       = localeutil.StringPhrase("problem.503")
	Problem503Detail = localeutil.StringPhrase("problem.503.detail")

	Problem504       = localeutil.StringPhrase("problem.504")
	Problem504Detail = localeutil.StringPhrase("problem.504.detail")

	Problem505       = localeutil.StringPhrase("problem.505")
	Problem505Detail = localeutil.StringPhrase("problem.505.detail")

	Problem506       = localeutil.StringPhrase("problem.506")
	Problem506Detail = localeutil.StringPhrase("problem.506.detail")

	Problem507       = localeutil.StringPhrase("problem.507")
	Problem507Detail = localeutil.StringPhrase("problem.507.detail")

	Problem508       = localeutil.StringPhrase("problem.508")
	Problem508Detail = localeutil.StringPhrase("problem.508.detail")

	Problem510       = localeutil.StringPhrase("problem.510")
	Problem510Detail = localeutil.StringPhrase("problem.510.detail")

	Problem511       = localeutil.StringPhrase("problem.511")
	Problem511Detail = localeutil.StringPhrase("problem.511.detail")
)
