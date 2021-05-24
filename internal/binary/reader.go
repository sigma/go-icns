// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package binary

import (
	"encoding/binary"
	"io"

	"yrh.dev/icns/internal/utils"
)

type Reader []byte

func (r *Reader) Uint32() uint32 {
	v := binary.BigEndian.Uint32(*r)
	*r = (*r)[4:]
	return v
}

func (r *Reader) Section(n int) *Reader {
	r2 := (*r)[:n]
	*r = (*r)[n:]
	return &r2
}

func (r *Reader) Read(p []byte) (int, error) {
	var err error
	nr := len(*r)
	n := utils.Min(nr, len(p))
	if n == nr {
		err = io.EOF
	}
	s := r.Section(n)
	for i := 0; i < n; i++ {
		p[i] = (*s)[i]
	}
	return n, err
}
