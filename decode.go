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
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
)

// modern icon support embedding either JPEG or PNG
func jpegOrPngDecode(r io.Reader, _ Resolution) (image.Image, string, error) {
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

// The image format looks like:
// - a RLE-encoded payload
// - that contains in order R bytes, G bytes and B bytes
// - alpha is packed separately in the mask file
func decodePack(r io.Reader, res Resolution) (image.Image, string, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	flat := rleUnpack(body)

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

func rleUnpack(p []byte) []byte {
	var res []byte
	pos := 0

	for {
		if pos >= len(p) {
			break
		}

		b := p[pos]
		if b < 0x80 {
			n := int(b) + 1
			res = append(res, p[pos+1:pos+1+n]...)
			pos += 1 + n
		} else {
			x := p[pos+1]
			n := int(b-0x80) + 3
			for i := 0; i < n; i++ {
				res = append(res, x)
			}
			pos += 2
		}
	}
	return res
}

// the separate mask file contains just the alpha channel
// is it optional?
func decodeAlpha(r io.Reader, res Resolution) (image.Image, string, error) {
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
