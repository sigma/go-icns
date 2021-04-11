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

package codec

import (
	"image"
	"io"
	"io/ioutil"

	"github.com/sigma/go-icns/internal/rle"
	"github.com/sigma/go-icns/internal/utils"
)

type argbCodec struct {
	header string
}

func (c *argbCodec) Encode(w io.Writer, img image.Image) error {
	if nrgba, ok := img.(*image.NRGBA); ok {
		w.Write([]byte(c.header))
		w.Write(rle.Encode(utils.NRGBAChannel(nrgba, 3)))
		for i := 0; i < 3; i++ {
			w.Write(rle.Encode(utils.NRGBAChannel(nrgba, i)))
		}
		return nil
	}
	return c.Encode(w, utils.Img2NRGBA(img))
}

func (c *argbCodec) Decode(r io.Reader, res Resolution) (image.Image, string, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	flat := rle.Decode(body[len(c.header):]) // skip header

	size := int(res * res)
	pixels := make([]byte, 4*size)
	for i := 0; i < size; i++ {
		pixels[i*4] = flat[size+i]
		pixels[i*4+1] = flat[2*size+i]
		pixels[i*4+2] = flat[3*size+i]
		pixels[i*4+3] = flat[i]
	}

	rect := image.Rect(0, 0, int(res), int(res))
	img := &image.NRGBA{
		Pix:    pixels,
		Stride: 4 * rect.Dx(),
		Rect:   rect,
	}
	return img, "argb", nil
}

var ARGBCodec = &argbCodec{
	header: "ARGB",
}
