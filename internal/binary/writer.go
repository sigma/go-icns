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

package binary

import "encoding/binary"

type Writer []byte

func (w *Writer) Uint32(v uint32) {
	binary.BigEndian.PutUint32(*w, v)
	*w = (*w)[4:]
}

func (w *Writer) Section(body []byte) {
	copy((*w), body)
	*w = (*w)[len(body):]
}
