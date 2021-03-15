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

package icns

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
)

type reader []byte

func (r *reader) uint8() uint8 {
	v := (*r)[0]
	*r = (*r)[1:]
	return v
}

func (r *reader) uint16() uint16 {
	v := binary.BigEndian.Uint16(*r)
	*r = (*r)[2:]
	return v
}

func (r *reader) uint32() uint32 {
	v := binary.BigEndian.Uint32(*r)
	*r = (*r)[4:]
	return v
}

func (r *reader) uint64() uint64 {
	v := binary.BigEndian.Uint64(*r)
	*r = (*r)[8:]
	return v
}

func (r *reader) sub(n int) *reader {
	b2 := (*r)[:n]
	*r = (*r)[n:]
	return &b2
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func (r *reader) Read(p []byte) (int, error) {
	n := min(len(*r), len(p))
	s := r.sub(n)
	for i := 0; i < n; i++ {
		p[i] = (*s)[i]
	}
	return n, nil
}

func readICNS(r reader) (*ICNS, error) {
	hdr := r.uint32()
	if hdr != magic {
		return nil, fmt.Errorf("wrong magic number for ICNS file: %x", hdr)
	}

	_ = r.uint32() // size

	minCompat := Newest
	maxCompat := Oldest

	var assets []*img
	for {
		if len(r) == 0 {
			break
		}

		code := r.uint32()
		size := int(r.uint32())
		sub := r.sub(size - 8) // size value includes both uint32 for code and size

		if f, ok := supportedFormats[code]; ok {
			i, err := f.decode(sub)
			if err != nil {
				continue
			}

			if f.compat < minCompat {
				minCompat = f.compat
			}

			if f.compat > maxCompat {
				maxCompat = f.compat
			}

			assets = append(assets, &img{
				Image:  i,
				format: f,
			})
		}
	}

	return &ICNS{
		minCompat: minCompat,
		maxCompat: maxCompat,
		assets:    assets,
	}, nil
}

// Decode loads a .icns file from the provided reader.
func Decode(r io.Reader) (*ICNS, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return readICNS(bytes)
}
