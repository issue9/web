// SPDX-License-Identifier: MIT

package main

import "github.com/issue9/web/micro"

func main() {
	s := micro.NewGateway(nil, nil)

	s.Serve()
}
