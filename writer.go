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

func (w *writer) uint8(v uint8) {
	(*w)[0] = v
	*w = (*w)[1:]
}

func (w *writer) uint16(v uint16) {
	binary.BigEndian.PutUint16(*w, v)
	*w = (*w)[2:]
}

func (w *writer) uint32(v uint32) {
	binary.BigEndian.PutUint32(*w, v)
	*w = (*w)[4:]
}

func (w *writer) uint64(v uint64) {
	binary.BigEndian.PutUint64(*w, v)
	*w = (*w)[8:]
}

func (w *writer) sub(body []byte) {
	for i, b := range body {
		(*w)[i] = b
	}
	*w = (*w)[len(body):]
}

// Encode writes a .icns file to the provided writer.
func Encode(w io.Writer, i *ICNS) error {
	n := len(i.assets)
	buffers := make([]bytes.Buffer, n)
	sizes := make([]uint32, n)
	types := make([]uint32, n)
	var totalSize uint32 = 8

	for idx, a := range i.assets {
		if err := a.format.encode(&buffers[idx], a.Image); err != nil {
			return err
		}
		types[idx] = a.format.code
		sizes[idx] = uint32(buffers[idx].Len()) + 8
		totalSize += sizes[idx]
	}

	data := make([]byte, totalSize)
	wd := writer(data)
	wd.uint32(magic)
	wd.uint32(totalSize)

	for idx := range buffers {
		wd.uint32(types[idx])
		wd.uint32(sizes[idx])
		wd.sub(buffers[idx].Bytes())
	}

	_, err := w.Write(data)
	return err
}
