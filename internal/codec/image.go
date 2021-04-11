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
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
)

type imageCodec struct{}

func (c *imageCodec) Encode(w io.Writer, img image.Image) error {
	// Unconditionally encode as PNG.
	return png.Encode(w, img)
}

func (c *imageCodec) Decode(r io.Reader, _ Resolution) (image.Image, string, error) {
	// we might have to re-read.
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, "", err
	}
	reader := bytes.NewReader(data)
	if img, err := jpeg.Decode(reader); err == nil {
		return img, "jpeg", nil
	}
	_, _ = reader.Seek(0, io.SeekStart)
	img, err := png.Decode(reader)
	if err != nil {
		return nil, "", err
	}
	return img, "png", nil
}

var ImageCodec = &imageCodec{}
