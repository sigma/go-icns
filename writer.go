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
	"io"

	"github.com/sigma/go-icns/internal/binary"
	"github.com/sigma/go-icns/internal/utils"
)

// Encode writes a .icns file to the provided writer.
func Encode(w io.Writer, i *ICNS) error {
	buffers := make([]*bytes.Buffer, 0)
	sizes := make([]uint32, 0)
	types := make([]uint32, 0)
	var totalSize uint32 = 8

	for _, a := range i.assets {
		encoder := a.format.codec.Encode
		if encoder == nil {
			continue
		}

		// encode mask first
		if a.format.combineCode != 0 {
			// the encoders expect an NRGBA instance
			a.Image = utils.Img2NRGBA(a.Image)

			// encode alpha channel as separated mask
			mformat := supportedMaskFormats[a.format.combineCode]
			buf := new(bytes.Buffer)
			if err := mformat.codec.Encode(buf, a.Image); err != nil {
				return err
			}
			size := uint32(buf.Len()) + 8
			buffers = append(buffers, buf)
			types = append(types, mformat.code)
			sizes = append(sizes, size)
			totalSize += size
		}

		buf := new(bytes.Buffer)
		if err := encoder(buf, a.Image); err != nil {
			return err
		}
		size := uint32(buf.Len()) + 8
		buffers = append(buffers, buf)
		types = append(types, a.format.code)
		sizes = append(sizes, size)
		totalSize += size
	}

	data := make([]byte, totalSize)
	wd := binary.Writer(data)
	wd.Uint32(magic)
	wd.Uint32(totalSize)

	for idx := range buffers {
		wd.Uint32(types[idx])
		wd.Uint32(sizes[idx])
		wd.Section(buffers[idx].Bytes())
	}

	_, err := w.Write(data)
	return err
}
