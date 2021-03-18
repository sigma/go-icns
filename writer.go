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
	"encoding/binary"
	"io"
)

type writer []byte

func (w *writer) uint32(v uint32) {
	binary.BigEndian.PutUint32(*w, v)
	*w = (*w)[4:]
}

func (w *writer) section(body []byte) {
	copy((*w), body)
	*w = (*w)[len(body):]
}

// Encode writes a .icns file to the provided writer.
func Encode(w io.Writer, i *ICNS) error {
	buffers := make([]*bytes.Buffer, 0)
	sizes := make([]uint32, 0)
	types := make([]uint32, 0)
	var totalSize uint32 = 8

	for _, a := range i.assets {
		encoder := a.format.encode
		if encoder == nil {
			continue
		}

		// encode mask first
		if a.format.combineCode != 0 {
			// the encoders expect an NRGBA instance
			a.Image = toNRGBA(a.Image)

			// encode alpha channel as separated mask
			mformat := supportedMaskFormats[a.format.combineCode]
			buf := new(bytes.Buffer)
			if err := mformat.encode(buf, a.Image); err != nil {
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
	wd := writer(data)
	wd.uint32(magic)
	wd.uint32(totalSize)

	for idx := range buffers {
		wd.uint32(types[idx])
		wd.uint32(sizes[idx])
		wd.section(buffers[idx].Bytes())
	}

	_, err := w.Write(data)
	return err
}
