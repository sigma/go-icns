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

import "github.com/sigma/go-icns/internal/codec"

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
