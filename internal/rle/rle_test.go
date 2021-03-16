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

package rle_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sigma/go-icns/internal/rle"
)

func TestRLE(t *testing.T) {
	data := []struct {
		name     string
		enc, dec []byte
	}{
		{
			"empty",
			nil,
			nil,
		},
		{
			"small",
			[]byte{0x02, 0x01, 0x02, 0x02, 0x80, 0x03, 0x81, 0x04, 0x82, 0x05},
			[]byte{0x01, 0x02, 0x02, 0x03, 0x03, 0x03, 0x04, 0x04, 0x04, 0x04, 0x05, 0x05, 0x05, 0x05, 0x05},
		},
		{
			"zeros overload",
			[]byte{0xff, 0x00, 0x82, 0x00},
			make([]byte, 135),
		},
	}

	for _, tt := range data {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			decoded := rle.Decode(tt.enc)
			if diff := cmp.Diff(tt.dec, decoded); diff != "" {
				t.Errorf("Decode() mismatch (-want +got):\n%s", diff)
			}

			encoded := rle.Encode(tt.dec)
			if diff := cmp.Diff(tt.enc, encoded); diff != "" {
				t.Errorf("Encode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
