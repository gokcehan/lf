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

type mapRange struct {
	first byte
	last  byte
	r     rune
}

var mapRanges = []mapRange{
	{0, 3, '\x00'},
	{4, 4, RuneError},
	{5, 5, '\x09'},
	{6, 6, RuneError},
	{7, 7, '\x7f'},
	{8, 10, RuneError},
	{11, 19, '\x0b'},
	{20, 20, RuneError},
	{21, 21, '\u0085'},
	{22, 22, '\x08'},
	{23, 23, RuneError},
	{24, 25, '\x18'},
	{26, 27, RuneError},
	{28, 31, '\x1c'},
	{32, 36, RuneError},
	{37, 37, '\x0a'},
	{38, 38, '\x17'},
	{39, 39, '\x1b'},
	{40, 44, RuneError},
	{45, 47, '\x05'},
	{48, 49, RuneError},
	{50, 50, '\x16'},
	{51, 54, RuneError},
	{55, 55, '\x04'},
	{56, 59, RuneError},
	{60, 61, '\x14'},
	{62, 62, RuneError},
	{63, 63, '\x1a'},
	{64, 64, '\x20'},
	{65, 65, '\xa0'},
	{66, 74, RuneError},
	{75, 75, '.'},
	{76, 76, '<'},
	{77, 77, '('},
	{78, 78, '+'},
	{79, 79, '|'},
	{80, 80, '&'},
	{81, 89, RuneError},
	{90, 90, '!'},
	{91, 91, '$'},
	{92, 92, '*'},
	{93, 93, ')'},
	{94, 94, ';'},
	{95, 95, '\u00ac'},
	{96, 96, '\x2d'},
	{97, 97, '\x2f'},
	{98, 105, RuneError},
	{106, 106, '\u00a6'},
	{107, 107, '\x2c'},
	{108, 108, '%'},
	{109, 109, '\x5f'},
	{110, 111, '\x3e'},
	{112, 120, RuneError},
	{121, 121, '\x60'},
	{122, 122, ':'},
	{123, 123, '#'},
	{124, 124, '@'},
	{125, 125, '\x27'},
	{126, 126, '='},
	{127, 127, '\x22'},
	{128, 128, RuneError},
	{129, 137, 'a'},
	{138, 142, RuneError},
	{143, 143, '\u00b1'},
	{144, 144, RuneError},
	{145, 153, 'j'},
	{154, 160, RuneError},
	{161, 161, '~'},
	{162, 169, 's'},
	{170, 175, RuneError},
	{176, 176, '\x5e'},
	{177, 185, RuneError},
	{186, 186, '['},
	{187, 187, ']'},
	{188, 191, RuneError},
	{192, 192, '{'},
	{193, 201, 'A'},
	{202, 202, '\u00ad'},
	{203, 207, RuneError},
	{208, 208, '}'},
	{209, 217, 'J'},
	{218, 223, RuneError},
	{224, 224, '\x5c'},
	{226, 233, 'S'},
	{234, 239, RuneError},
	{240, 249, '0'},
	{250, 255, RuneError},
}

func TestEBCDICvsASCII(t *testing.T) {
	t.Logf("EBCDIC/ASCII identity values")
	for _, rng := range mapRanges {
		if rng.r == RuneError {
			continue
		}
		for j := 0; j <= int(rng.last-rng.first); j++ {
			b := rng.first + byte(j)
			r := rng.r + rune(j)
			verifyMap(t, EBCDIC, b, r)
		}
	}
}
func TestEBDICvsInvalid(t *testing.T) {
	t.Logf("EBCDIC invalid transforms")
	for _, rng := range mapRanges {
		if rng.r != RuneError {
			continue
		}
		for j := 0; j <= int(rng.last-rng.first); j++ {
			b := rng.first + byte(j)
			verifyToUTF(t, EBCDIC, b, RuneError)
		}
	}
}

func TestEBDICvsLargeUTF(t *testing.T) {
	t.Logf("Large UTF maps to subst char")
	verifyFromUTF(t, EBCDIC, '\x3F', '㿿')
}
