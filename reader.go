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
	"fmt"
	"image"
	"image/draw"
	"io"
	"io/ioutil"

	"yrh.dev/icns/internal/binary"
)

func readICNS(r binary.Reader, metaOnly bool) (*ICNS, error) {
	hdr := r.Uint32()
	if hdr != magic {
		return nil, fmt.Errorf("wrong magic number for ICNS file: %x", hdr)
	}

	_ = r.Uint32() // size

	minCompat := Newest
	maxCompat := Oldest

	var assets []*img
	masks := make(map[uint32]image.Image)

	var unsupportedCodes []uint32
	for {
		if len(r) == 0 {
			break
		}

		code := r.Uint32()
		size := int(r.Uint32())
		sub := r.Section(size - 8) // size value includes both uint32 for code and size

		if f, ok := supportedMaskFormats[code]; ok {
			if metaOnly {
				continue
			}

			i, _, err := f.codec.Decode(sub, f.res)
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
			asset := &img{
				format: f,
			}

			if !metaOnly {
				i, enc, err := f.codec.Decode(sub, f.res)
				if err != nil {
					continue
				}

				// TODO: don't assume the mask is parsed first
				if m := masks[f.combineCode]; m != nil {
					r := image.Rect(0, 0, int(f.res), int(f.res))

					c := image.NewRGBA(r)

					draw.DrawMask(c, r, i, image.Pt(0, 0), m, image.Pt(0, 0), draw.Over)
					i = c
				}

				asset.Image = i
				asset.encoder = enc
			}

			assets = append(assets, asset)

			if f.compat < minCompat {
				minCompat = f.compat
			}

			if f.compat > maxCompat {
				maxCompat = f.compat
			}

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
	return readICNS(bytes, false)
}
