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

	"github.com/sigma/go-icns/internal/utils"
)

type maskCodec struct{}

func (c *maskCodec) Encode(w io.Writer, img image.Image) error {
	if nrgba, ok := img.(*image.NRGBA); ok {
		alpha := utils.NRGBAChannel(nrgba, 3)
		w.Write(alpha)
		return nil
	}
	return c.Encode(w, utils.Img2NRGBA(img))
}

func (c *maskCodec) Decode(r io.Reader, res Resolution) (image.Image, string, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	rect := image.Rect(0, 0, int(res), int(res))
	img := &image.Alpha{
		Pix:    body,
		Stride: 1 * rect.Dx(),
		Rect:   rect,
	}
	return img, "mask", nil
}

var MaskCodec = &maskCodec{}
