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

type packCodec struct{}

func (c *packCodec) Encode(w io.Writer, img image.Image) error {
	if nrgba, ok := img.(*image.NRGBA); ok {
		for i := 0; i < 3; i++ {
			c := utils.NRGBAChannel(nrgba, i)
			w.Write(rle.Encode(c))
		}
		return nil
	}
	return c.Encode(w, utils.Img2NRGBA(img))
}

func (c *packCodec) Decode(r io.Reader, res Resolution) (image.Image, string, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	flat := rle.Decode(body)

	size := int(res * res)
	pixels := make([]byte, 4*size)
	for i := 0; i < size; i++ {
		pixels[i*4] = flat[i]
		pixels[i*4+1] = flat[size+i]
		pixels[i*4+2] = flat[2*size+i]
		pixels[i*4+3] = 0xff
	}

	rect := image.Rect(0, 0, int(res), int(res))
	img := &image.NRGBA{
		Pix:    pixels,
		Stride: 4 * rect.Dx(),
		Rect:   rect,
	}
	return img, "icon", nil
}

var PackCodec = &packCodec{}
