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
	"image"
	"io"
	"io/ioutil"
	"path"
	"testing"
)

type test interface {
	Helper()
	Fatal(args ...interface{})
}

func testdataFileReader(t test, fname string) *bytes.Reader {
	t.Helper()

	body, err := ioutil.ReadFile(path.Join("testdata", fname))
	if err != nil {
		t.Fatal(err)
	}

	return bytes.NewReader(body)
}

func TestDecode(t *testing.T) {
	t.Parallel()
	img, fmt, err := image.Decode(testdataFileReader(t, "mit.icns"))
	if err != nil {
		t.Fatal(err)
	}

	if fmt != "icns" {
		t.Errorf("unexpected image format: got %s, want icns", fmt)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 1024 || bounds.Dy() != 1024 {
		t.Errorf("unexpected image size: got %dx%d, want 1024x1024", bounds.Dx(), bounds.Dy())
	}
}

func TestDecodeConfig(t *testing.T) {
	t.Parallel()
	cfg, fmt, err := image.DecodeConfig(testdataFileReader(t, "mit.icns"))
	if err != nil {
		t.Fatal(err)
	}

	if fmt != "icns" {
		t.Errorf("unexpected image format: got %s, want icns", fmt)
	}

	if cfg.Width != 1024 || cfg.Height != 1024 {
		t.Errorf("unexpected image size: got %dx%d, want 1024x1024", cfg.Width, cfg.Height)
	}
}

func BenchmarkDecodeConfig(b *testing.B) {
	r := testdataFileReader(b, "mit.icns")
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			b.Fatal(err)
		}
		if _, _, err := image.DecodeConfig(r); err != nil {
			b.Fatal(err)
		}
	}
}
