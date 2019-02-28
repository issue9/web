module github.com/issue9/web

require (
	github.com/caixw/gobuild v0.7.2
	github.com/issue9/assert v1.3.2
	github.com/issue9/is v1.1.3
	github.com/issue9/logs/v2 v2.2.2
	github.com/issue9/middleware v1.5.3
	github.com/issue9/mux/v2 v2.1.1
	github.com/issue9/query v1.0.1
	github.com/issue9/upload v1.1.2
	github.com/issue9/utils v1.0.2
	github.com/issue9/version v1.0.2
	golang.org/x/text v0.3.0
	gopkg.in/yaml.v2 v2.2.1
)

replace (
	golang.org/x/sys => github.com/golang/sys v0.0.0-20180905080454-ebe1bf3edb33
	golang.org/x/text v0.3.0 => github.com/golang/text v0.3.0
)
