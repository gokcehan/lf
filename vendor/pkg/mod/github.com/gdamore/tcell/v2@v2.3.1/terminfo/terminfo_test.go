// Copyright 2021 The TCell Authors
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

package terminfo

import (
	"bytes"
	"testing"
	"time"
)

// This terminfo entry is a stripped down version from
// xterm-256color, but I've added some of my own entries.
var testTerminfo = &Terminfo{
	Name:      "simulation_test",
	Columns:   80,
	Lines:     24,
	Colors:    256,
	Bell:      "\a",
	Blink:     "\x1b2ms$<20>something",
	Reverse:   "\x1b[7m",
	SetFg:     "\x1b[%?%p1%{8}%<%t3%p1%d%e%p1%{16}%<%t9%p1%{8}%-%d%e38;5;%p1%d%;m",
	SetBg:     "\x1b[%?%p1%{8}%<%t4%p1%d%e%p1%{16}%<%t10%p1%{8}%-%d%e48;5;%p1%d%;m",
	AltChars:  "``aaffggiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz{{||}}~~",
	Mouse:     "\x1b[M",
	SetCursor: "\x1b[%i%p1%d;%p2%dH",
	PadChar:   "\x00",
}

func TestTerminfoExpansion(t *testing.T) {

	ti := testTerminfo

	// Tests %i and basic parameter strings too
	if ti.TGoto(7, 9) != "\x1b[10;8H" {
		t.Error("TGoto expansion failed")
	}

	// This tests some conditionals
	if ti.TParm("A[%p1%2.2X]B", 47) != "A[2F]B" {
		t.Error("TParm conditionals failed")
	}

	// Color tests.
	if ti.TParm(ti.SetFg, 7) != "\x1b[37m" {
		t.Error("SetFg(7) failed")
	}
	if ti.TParm(ti.SetFg, 15) != "\x1b[97m" {
		t.Error("SetFg(15) failed")
	}
	if ti.TParm(ti.SetFg, 200) != "\x1b[38;5;200m" {
		t.Error("SetFg(200) failed")
	}
}

func TestTerminfoDelay(t *testing.T) {
	ti := testTerminfo
	buf := bytes.NewBuffer(nil)
	now := time.Now()
	ti.TPuts(buf, ti.Blink)
	then := time.Now()
	s := string(buf.Bytes())
	if s != "\x1b2mssomething" {
		t.Errorf("Terminfo delay failed: %s", s)
	}
	if then.Sub(now) < time.Millisecond*20 {
		t.Error("Too short delay")
	}
	if then.Sub(now) > time.Millisecond*50 {
		t.Error("Too late delay")
	}
}

func BenchmarkSetFgBg(b *testing.B) {
	ti := testTerminfo

	for i := 0; i < b.N; i++ {
		ti.TParm(ti.SetFg, 100, 200)
		ti.TParm(ti.SetBg, 100, 200)
	}
}
