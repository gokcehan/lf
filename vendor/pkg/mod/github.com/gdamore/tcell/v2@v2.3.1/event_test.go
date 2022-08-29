// Copyright 2018 The TCell Authors
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

package tcell

import (
	"testing"
	"time"
)

func eventLoop(s SimulationScreen, evch chan Event) {
	for {
		ev := s.PollEvent()
		if ev == nil {
			close(evch)
			return
		}
		select {
		case evch <- ev:
		case <-time.After(time.Second):
		}
	}
}

func TestMouseEvents(t *testing.T) {

	s := mkTestScreen(t, "")
	defer s.Fini()

	s.EnableMouse()
	s.InjectMouse(4, 9, Button1, ModCtrl)
	evch := make(chan Event)
	em := &EventMouse{}
	done := false
	go eventLoop(s, evch)

	for !done {
		select {
		case ev := <-evch:
			if evm, ok := ev.(*EventMouse); ok {
				em = evm
				done = true
			}
			continue
		case <-time.After(time.Second):
			done = true
		}
	}

	if x, y := em.Position(); x != 4 || y != 9 {
		t.Errorf("Mouse position wrong (%v, %v)", x, y)
	}
	if em.Buttons() != Button1 {
		t.Errorf("Should be Button1")
	}
	if em.Modifiers() != ModCtrl {
		t.Errorf("Modifiers should be control")
	}
}
