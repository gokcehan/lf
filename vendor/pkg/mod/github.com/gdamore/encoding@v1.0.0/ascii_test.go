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
	"testing"
)

func TestASCII(t *testing.T) {
	t.Logf("ASCII identity transforms")
	for i := 0; i < 128; i++ {
		verifyMap(t, ASCII, byte(i), rune(i))
	}

	t.Logf("High order bytes map to RuneError")
	for i := 128; i < 256; i++ {
		verifyToUTF(t, ASCII, byte(i), RuneError)
	}

	t.Logf("High order UTF maps to ASCIISub")
	for i := 128; i < 256; i++ {
		verifyFromUTF(t, ASCII, ASCIISub, rune(i))
	}

	t.Logf("Large UTF maps to ASCIISub")
	verifyFromUTF(t, ASCII, ASCIISub, '㿿')
}
