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

// Package icns provides read/write operations for the Apple ICNS file format.
// It currently only supports a subset of the specification, covering JPEG and PNG data types.
package icns

import (
	"bytes"
	"fmt"
	"image"
	"io"

	"github.com/sigma/go-icns/internal/codec"
)

const (
	magic uint32 = ('i'<<24 | 'c'<<16 | 'n'<<8 | 's')
	is32  uint32 = ('i'<<24 | 's'<<16 | '3'<<8 | '2')
	s8mk  uint32 = ('s'<<24 | '8'<<16 | 'm'<<8 | 'k')
	il32  uint32 = ('i'<<24 | 'l'<<16 | '3'<<8 | '2')
	l8mk  uint32 = ('l'<<24 | '8'<<16 | 'm'<<8 | 'k')
	ih32  uint32 = ('i'<<24 | 'h'<<16 | '3'<<8 | '2')
	h8mk  uint32 = ('h'<<24 | '8'<<16 | 'm'<<8 | 'k')
	it32  uint32 = ('i'<<24 | 't'<<16 | '3'<<8 | '2')
	t8mk  uint32 = ('t'<<24 | '8'<<16 | 'm'<<8 | 'k')
	icp4  uint32 = ('i'<<24 | 'c'<<16 | 'p'<<8 | '4')
	icp5  uint32 = ('i'<<24 | 'c'<<16 | 'p'<<8 | '5')
	icp6  uint32 = ('i'<<24 | 'c'<<16 | 'p'<<8 | '6')
	ic04  uint32 = ('i'<<24 | 'c'<<16 | '0'<<8 | '4')
	ic05  uint32 = ('i'<<24 | 'c'<<16 | '0'<<8 | '5')
	ic07  uint32 = ('i'<<24 | 'c'<<16 | '0'<<8 | '7')
	ic08  uint32 = ('i'<<24 | 'c'<<16 | '0'<<8 | '8')
	ic09  uint32 = ('i'<<24 | 'c'<<16 | '0'<<8 | '9')
	ic10  uint32 = ('i'<<24 | 'c'<<16 | '1'<<8 | '0')
	ic11  uint32 = ('i'<<24 | 'c'<<16 | '1'<<8 | '1')
	ic12  uint32 = ('i'<<24 | 'c'<<16 | '1'<<8 | '2')
	ic13  uint32 = ('i'<<24 | 'c'<<16 | '1'<<8 | '3')
	ic14  uint32 = ('i'<<24 | 'c'<<16 | '1'<<8 | '4')
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

// Resolution represents the supported resolutions in pixels.
type Resolution = codec.Resolution

// All supported resolutions
const (
	Pixel16   Resolution = 16
	Pixel32   Resolution = 32
	Pixel48   Resolution = 48
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
	code        uint32
	combineCode uint32
	res         Resolution
	compat      Compatibility
	codec       codec.Codec
}

var (
	supportedImageFormats map[uint32]*format
	supportedMaskFormats  map[uint32]*format
)

func init() {
	supportedImageFormats = make(map[uint32]*format)
	supportedMaskFormats = make(map[uint32]*format)

	legacyFormats := []struct {
		code uint32
		mask uint32
		res  Resolution
	}{
		{is32, s8mk, Pixel16},
		{il32, l8mk, Pixel32},
		{ih32, h8mk, Pixel16},
		{it32, t8mk, Pixel32},
	}

	for _, f := range legacyFormats {
		supportedImageFormats[f.code] = &format{
			code:        f.code,
			combineCode: f.mask,
			res:         f.res,
			compat:      Allegro,
			codec:       codec.PackCodec,
		}

		supportedMaskFormats[f.mask] = &format{
			code:        f.mask,
			combineCode: f.code,
			res:         f.res,
			compat:      Allegro,
			codec:       codec.MaskCodec,
		}
	}

	argbFormats := []struct {
		code uint32
		res  Resolution
	}{
		{ic04, Pixel16},
		{ic05, Pixel32},
	}

	for _, f := range argbFormats {
		supportedImageFormats[f.code] = &format{
			code:   f.code,
			res:    f.res,
			compat: Cheetah, // not quite sure
			codec:  codec.ARGBCodec,
		}
	}

	modernFormats := []struct {
		code   uint32
		res    Resolution
		compat Compatibility
	}{
		{icp4, Pixel16, Lion},
		{icp5, Pixel32, Lion},
		{icp6, Pixel64, Lion},
		{ic07, Pixel128, Lion},
		{ic08, Pixel256, Leopard},
		{ic09, Pixel512, Leopard},
		{ic10, Pixel1024, Lion},
		{ic11, Pixel32, MountainLion},
		{ic12, Pixel64, MountainLion},
		{ic13, Pixel256, MountainLion},
		{ic14, Pixel512, MountainLion},
	}

	for _, f := range modernFormats {
		supportedImageFormats[f.code] = &format{
			code:   f.code,
			res:    f.res,
			compat: f.compat,
			codec:  codec.ImageCodec,
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
	format  *format
	encoder string
}

// ICNS encapsulates the Applie Icon Image format specification.
type ICNS struct {
	minCompat, maxCompat Compatibility
	assets               []*img
	unsupportedCodes     []uint32
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
	for _, f := range supportedImageFormats {
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

// Info provides information about the ICNS
func (i *ICNS) Info() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%d images:\n", len(i.assets)+len(i.unsupportedCodes))
	for _, a := range i.assets {
		fmt.Fprintf(buf, "[%s] %s image with resolution %d\n", codeRepr(a.format.code), a.encoder, a.Image.Bounds().Dx())
	}
	for _, c := range i.unsupportedCodes {
		fmt.Fprintf(buf, "[%s] unsupported image format\n", codeRepr(c))
	}
	return buf.String()
}
