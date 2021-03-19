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
	"image"
	"image/draw"
	"io"

	"github.com/sigma/go-icns/internal/rle"
)

func toNRGBA(img image.Image) *image.NRGBA {
	r := img.Bounds()
	res := image.NewNRGBA(r)
	draw.Draw(res, r, img, image.Point{}, draw.Over)
	return res
}

func nrgbaChannel(img *image.NRGBA, c int) []byte {
	size := len(img.Pix) / 4
	res := make([]byte, size)

	for idx := range res {
		res[idx] = img.Pix[idx*4+c]
	}
	return res
}

func encodePack(w io.Writer, img image.Image) error {
	if nrgba, ok := img.(*image.NRGBA); ok {
		for i := 0; i < 3; i++ {
			c := nrgbaChannel(nrgba, i)
			w.Write(rle.Encode(c))
		}
		return nil
	}
	return encodePack(w, toNRGBA(img))
}

func encodeMask(w io.Writer, img image.Image) error {
	if nrgba, ok := img.(*image.NRGBA); ok {
		alpha := nrgbaChannel(nrgba, 3)
		w.Write(alpha)
		return nil
	}
	return encodeMask(w, toNRGBA(img))
}

func encodeARGB(w io.Writer, img image.Image) error {
	if nrgba, ok := img.(*image.NRGBA); ok {
		w.Write([]byte("ARGB"))
		w.Write(rle.Encode(nrgbaChannel(nrgba, 3)))
		for i := 0; i < 3; i++ {
			w.Write(rle.Encode(nrgbaChannel(nrgba, i)))
		}
		return nil
	}
	return encodeARGB(w, toNRGBA(img))
}
