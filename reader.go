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
	"image"
	"image/draw"
	"io"
	"io/ioutil"
)

type reader []byte

func (r *reader) uint32() uint32 {
	v := binary.BigEndian.Uint32(*r)
	*r = (*r)[4:]
	return v
}

func (r *reader) section(n int) *reader {
	r2 := (*r)[:n]
	*r = (*r)[n:]
	return &r2
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func (r *reader) Read(p []byte) (int, error) {
	var err error
	nr := len(*r)
	n := min(nr, len(p))
	if n == nr {
		err = io.EOF
	}
	s := r.section(n)
	for i := 0; i < n; i++ {
		p[i] = (*s)[i]
	}
	return n, err
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
	masks := make(map[uint32]image.Image)

	var unsupportedCodes []uint32
	for {
		if len(r) == 0 {
			break
		}

		code := r.uint32()
		size := int(r.uint32())
		sub := r.section(size - 8) // size value includes both uint32 for code and size

		if f, ok := supportedMaskFormats[code]; ok {
			i, _, err := f.decode(sub, f.res)
			if err != nil {
				continue
			}

			if f.compat < minCompat {
				minCompat = f.compat
			}

			if f.compat > maxCompat {
				maxCompat = f.compat
			}

			masks[code] = i

			continue
		}

		if f, ok := supportedImageFormats[code]; ok {
			i, enc, err := f.decode(sub, f.res)
			if err != nil {
				continue
			}

			if f.compat < minCompat {
				minCompat = f.compat
			}

			if f.compat > maxCompat {
				maxCompat = f.compat
			}

			// TODO: don't assume the mask is parsed first
			if m := masks[f.combineCode]; m != nil {
				r := image.Rect(0, 0, int(f.res), int(f.res))

				c := image.NewRGBA(r)

				draw.DrawMask(c, r, i, image.Pt(0, 0), m, image.Pt(0, 0), draw.Over)
				i = c
			}

			assets = append(assets, &img{
				Image:   i,
				format:  f,
				encoder: enc,
			})

			continue
		}

		unsupportedCodes = append(unsupportedCodes, code)
	}

	return &ICNS{
		minCompat:        minCompat,
		maxCompat:        maxCompat,
		assets:           assets,
		unsupportedCodes: unsupportedCodes,
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
