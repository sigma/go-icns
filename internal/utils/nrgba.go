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

package utils

import (
	"image"
	"image/draw"
)

func Img2NRGBA(img image.Image) *image.NRGBA {
	r := img.Bounds()
	res := image.NewNRGBA(r)
	draw.Draw(res, r, img, image.Point{}, draw.Over)
	return res
}

func NRGBAChannel(img *image.NRGBA, c int) []byte {
	size := len(img.Pix) / 4
	res := make([]byte, size)

	for idx := range res {
		res[idx] = img.Pix[idx*4+c]
	}
	return res
}
