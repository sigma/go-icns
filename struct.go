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
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

const (
	magic uint32 = 0x69636e73
)

// Resolution represents the supported resolutions in pixels.
type Resolution uint

// All supported resolutions
const (
	Pixel16   Resolution = 16
	Pixel32   Resolution = 32
	Pixel64   Resolution = 64
	Pixel128  Resolution = 128
	Pixel256  Resolution = 256
	Pixel512  Resolution = 512
	Pixel1024 Resolution = 1024
)

// Compatibility represents compatibility with an OS version.
type Compatibility uint

const (
	// Allegro is 8.5
	Allegro Compatibility = iota
	// Cheetah is 10.0
	Cheetah
	// Leopard is 10.5
	Leopard
	// Lion is 10.7
	Lion
	// MountainLion is 10.8
	MountainLion
	// Newest version
	Newest = MountainLion
	// Oldest version
	Oldest Compatibility = Allegro
)

type format struct {
	code   uint32
	res    Resolution
	compat Compatibility
	encode func(io.Writer, image.Image) error
	decode func(io.Reader) (image.Image, error)
}

const (
	ic07 uint32 = ('i'<<24 | 'c'<<16 | '0'<<8 | '7')
	ic08 uint32 = ('i'<<24 | 'c'<<16 | '0'<<8 | '8')
	ic09 uint32 = ('i'<<24 | 'c'<<16 | '0'<<8 | '9')
	ic10 uint32 = ('i'<<24 | 'c'<<16 | '1'<<8 | '0')
	ic11 uint32 = ('i'<<24 | 'c'<<16 | '1'<<8 | '1')
	ic12 uint32 = ('i'<<24 | 'c'<<16 | '1'<<8 | '2')
	ic13 uint32 = ('i'<<24 | 'c'<<16 | '1'<<8 | '3')
	ic14 uint32 = ('i'<<24 | 'c'<<16 | '1'<<8 | '4')
)

func codeRepr(c uint32) string {
	r := []rune{
		rune(c >> 24 & 0xff),
		rune(c >> 16 & 0xff),
		rune(c >> 8 & 0xff),
		rune(c & 0xff),
	}
	return string(r)
}

var supportedFormats map[uint32]*format

func jpegOrPngDecode(r io.Reader) (image.Image, error) {
	// we might have to re-read.
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)

	if img, err := jpeg.Decode(tee); err == nil {
		return img, nil
	}
	return png.Decode(&buf)
}

func init() {
	supportedFormats = make(map[uint32]*format)

	// TODO: support more legacy formats.
	// is32, s8mk, il32 and l8mk are still in use for example.
	modernFormats := []struct {
		code   uint32
		res    Resolution
		compat Compatibility
	}{
		{ic07, 128, Lion},
		{ic08, 256, Leopard},
		{ic09, 512, Leopard},
		{ic10, 1024, Lion},
		{ic11, 32, MountainLion},
		{ic12, 64, MountainLion},
		{ic13, 256, MountainLion},
		{ic14, 512, MountainLion},
	}

	for _, f := range modernFormats {
		supportedFormats[f.code] = &format{
			code:   f.code,
			res:    f.res,
			compat: f.compat,
			// always encode as PNG
			encode: png.Encode,
			// these can be either JPEG or PNG
			decode: png.Decode,
		}
	}

	// register into image decoding library. Use the highest available resolution for that purpose.
	image.RegisterFormat("icns", codeRepr(magic),
		func(r io.Reader) (image.Image, error) {
			i, err := Decode(r)
			if err != nil {
				return nil, err
			}
			return i.HighestResolution()
		},
		func(r io.Reader) (image.Config, error) {
			i, err := Decode(r)
			if err != nil {
				return image.Config{}, err
			}
			img, err := i.HighestResolution()
			if err != nil {
				return image.Config{}, err
			}
			return image.Config{
				ColorModel: img.ColorModel(),
				Width:      img.Bounds().Dx(),
				Height:     img.Bounds().Dy(),
			}, nil
		})
}

type img struct {
	image.Image
	format *format
}

// ICNS encapsulates the Applie Icon Image format specification.
type ICNS struct {
	minCompat, maxCompat Compatibility
	assets               []*img
}

// Option is the type for ICNS creation options.
type Option func(*ICNS)

// WithMinCompatibility sets the minimum expected compatibility (defaults to Oldest).
func WithMinCompatibility(c Compatibility) Option {
	return func(i *ICNS) {
		i.minCompat = c
	}
}

// WithMaxCompatibility sets the maximum expected compatibility (defaults to Newest).
func WithMaxCompatibility(c Compatibility) Option {
	return func(i *ICNS) {
		i.maxCompat = c
	}
}

// NewICNS creates a new icon based on provided options.
func NewICNS(opts ...Option) *ICNS {
	i := &ICNS{
		minCompat: Oldest,
		maxCompat: Newest,
	}

	for _, o := range opts {
		o(i)
	}

	return i
}

// ByResolution extracts an image from the icon, at the provided resolution.
func (i *ICNS) ByResolution(r Resolution) (image.Image, error) {
	for _, a := range i.assets {
		if a.format.res == r {
			return a.Image, nil
		}
	}
	return nil, fmt.Errorf("no image by that resolution")
}

// HighestResolution extracts the image from the icon that has the highest resolution.
func (i *ICNS) HighestResolution() (image.Image, error) {
	var res Resolution
	var img image.Image
	for _, a := range i.assets {
		if a.format.res > res {
			res = a.format.res
			img = a.Image
		}
	}

	if img == nil {
		return nil, fmt.Errorf("no valid image")
	}
	return img, nil
}

// Add adds new image to the icon, assuming its resolution is acceptable.
// This also replaces previous images at that resolution.
func (i *ICNS) Add(im image.Image) error {
	dx := im.Bounds().Dx()
	dy := im.Bounds().Dy()

	if dx != dy {
		return fmt.Errorf("image is not a square")
	}

	var supported bool
	for _, f := range supportedFormats {
		if f.compat < i.minCompat || f.compat > i.maxCompat {
			continue
		}

		if f.res == Resolution(dx) {
			supported = true

			var found bool
			for _, a := range i.assets {
				if a.format == f {
					found = true
					a.Image = im
				}
			}

			if !found {
				i.assets = append(i.assets, &img{
					Image:  im,
					format: f,
				})
			}
		}
	}

	if !supported {
		return fmt.Errorf("no available format for resolution %d", dx)
	}

	return nil
}
