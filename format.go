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
	"image/color"
	"io"
	"io/ioutil"

	"yrh.dev/icns/internal/codec"
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
			bytes, err := ioutil.ReadAll(r)
			if err != nil {
				return image.Config{}, err
			}
			i, err := readICNS(bytes, true)
			if err != nil {
				return image.Config{}, err
			}
			img, err := i.highestResolutionAsset()
			if err != nil {
				return image.Config{}, err
			}
			return image.Config{
				ColorModel: color.NRGBAModel,
				Width:      int(img.format.res),
				Height:     int(img.format.res),
			}, nil
		})
}
