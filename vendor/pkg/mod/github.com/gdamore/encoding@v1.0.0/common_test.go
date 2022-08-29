// Copyright 2018 Garrett D'Amore
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package encoding

import (
	"bytes"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/encoding"
)

func verifyMap(t *testing.T, enc encoding.Encoding, b byte, r rune) {
	verifyFromUTF(t, enc, b, r)
	verifyToUTF(t, enc, b, r)
}

func verifyFromUTF(t *testing.T, enc encoding.Encoding, b byte, r rune) {

	encoder := enc.NewEncoder()

	out := make([]byte, 6)
	utf := make([]byte, utf8.RuneLen(r))
	utf8.EncodeRune(utf, r)

	ndst, nsrc, err := encoder.Transform(out, utf, true)
	if err != nil {
		t.Errorf("Transform failed: %v", err)
	}
	if nsrc != len(utf) {
		t.Errorf("Length of source incorrect: %d != %d", nsrc, len(utf))
	}
	if ndst != 1 {
		t.Errorf("Dest length (%d) != 1", ndst)
	}
	if b != out[0] {
		t.Errorf("From UTF incorrect map %v != %v", b, out[0])
	}
}

func verifyToUTF(t *testing.T, enc encoding.Encoding, b byte, r rune) {
	decoder := enc.NewDecoder()

	out := make([]byte, 6)
	nat := []byte{b}
	utf := make([]byte, utf8.RuneLen(r))
	utf8.EncodeRune(utf, r)

	ndst, nsrc, err := decoder.Transform(out, nat, true)
	if err != nil {
		t.Errorf("Transform failed: %v", err)
	}
	if nsrc != 1 {
		t.Errorf("Src length (%d) != 1", nsrc)
	}
	if !bytes.Equal(utf, out[:ndst]) {
		t.Errorf("UTF expected %v, but got %v for %x\n", utf, out, b)
	}
}
