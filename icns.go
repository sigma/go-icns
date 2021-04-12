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
)

type img struct {
	image.Image
	format  *format
	encoder string
}

// ICNS encapsulates the Apple Icon Image format specification.
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

func (i *ICNS) highestResolutionAsset() (*img, error) {
	var res Resolution
	var img *img
	for _, a := range i.assets {
		if a.format.res > res {
			res = a.format.res
			img = a
		}
	}

	if img == nil {
		return nil, fmt.Errorf("no valid image")
	}
	return img, nil
}

// HighestResolution extracts the image from the icon that has the highest resolution.
func (i *ICNS) HighestResolution() (image.Image, error) {
	img, err := i.highestResolutionAsset()
	if err != nil {
		return nil, err
	}

	return img.Image, nil
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
