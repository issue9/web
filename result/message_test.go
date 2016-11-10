// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"testing"

	"github.com/issue9/assert"
)

func TestMessages(t *testing.T) {
	a := assert.New(t)

	status := 400
	a.Equal(status*scale, RegisterMessage(status, "status:400-0")).
		Equal("status:400-0", Message(status*scale)).
		Equal(len(messages), 1).
		Equal(len(indexes), 1)

	a.Equal(status*scale+1, RegisterMessage(status, "status:400-1")).
		Equal("status:400-1", Message(status*scale+1)).
		Equal(len(messages), 2).
		Equal(len(indexes), 1)

	status = 500
	a.Equal(status*scale, RegisterMessage(status, "status:500-0")).
		Equal("status:500-0", Message(status*scale)).
		Equal(len(messages), 3).
		Equal(len(indexes), 2)

	status = 400
	a.Equal(status*scale+2, RegisterMessage(status, "status:400-2")).
		Equal("status:400-2", Message(status*scale+2)).
		Equal(len(messages), 4).
		Equal(len(indexes), 2)

	//
	a.Equal(Message(-1000), codeNotExists)
}
