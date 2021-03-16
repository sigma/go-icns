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

// Package rle implements a simple RLE encoder, as used in ICNS format.
//
// The encoded format can be seen as a sequence of segments of the following form:
// - 1 control byte N
// - a number of data bytes
// If the control byte value is N < 0x80, then the decoded data contains the
//   N+1 encoded bytes as is
// If the control byte value is N >= 0x80, then the decoded data contains
//   N-0x80+3 repetitions of the following encoded byte
// In particular, this means that the second case can cover only repetitions of
//   3 or more bytes.
// Similarly, a "raw" sequence can contain only 128 (0x7f+1) bytes, and needs to be split
// if a longer non-repetitive pattern is seen.
package rle

type byteRec struct {
	b byte
	n int
}

// Encode RLE-encodes the provided bytes.
func Encode(b []byte) []byte {
	var res []byte

	if len(b) == 0 {
		return res
	}

	var records []*byteRec

	cur := &byteRec{
		b: b[0],
		n: 1,
	}

	for i := 1; i < len(b); i++ {
		c := b[i]
		if c != cur.b {
			records = append(records, cur)
			cur = &byteRec{
				b: c,
				n: 1,
			}
		} else {
			cur.n++
		}
	}
	records = append(records, cur)

	n := 0
	tmp := make([]byte, 0)

	flush := func() {
		if n == 0 {
			return
		}
		res = append(res, byte(n-1))
		res = append(res, tmp...)
		tmp = make([]byte, 0)
		n = 0
	}

	for _, r := range records {
		if r.n < 3 {
			if n+r.n <= 128 { // so the max segment length is 127 (0x7f)
				n += r.n
			} else {
				flush()
				n = r.n
			}
			for i := 0; i < r.n; i++ {
				tmp = append(tmp, r.b)
			}
		} else {
			flush()
			for r.n > 0 {
				// because we only compress sequences of 3+ characters
				// we encode repetitions of 3 to 130 as 0x80 to 0xff
				n := min(r.n, 130)
				res = append(res, byte(0x80+n-3), r.b)
				r.n -= n
			}
		}
	}
	flush()
	return res
}

// Decode RLE-decodes the provided bytes.
func Decode(p []byte) []byte {
	var res []byte
	pos := 0

	for {
		if pos >= len(p) {
			break
		}

		b := p[pos]
		if b < 0x80 {
			n := int(b) + 1
			res = append(res, p[pos+1:pos+1+n]...)
			pos += 1 + n
		} else {
			x := p[pos+1]
			n := int(b-0x80) + 3
			for i := 0; i < n; i++ {
				res = append(res, x)
			}
			pos += 2
		}
	}
	return res
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}
