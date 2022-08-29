// +build ignore

// Copyright 2015 The Tops'l Authors
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

package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type boxL struct {
	views.BoxLayout
}

var box = &boxL{}
var app = &views.Application{}

func (m *boxL) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			app.Quit()
			return true
		}
	}
	return m.BoxLayout.HandleEvent(ev)
}

func main() {

	title := views.NewTextBar()
	title.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorYellow))
	title.SetCenter("Vertical Layout", tcell.StyleDefault)
	top := views.NewText()
	mid := views.NewText()
	bot := views.NewText()

	top.SetText("Top-Right (0.0)\nLine Two")
	mid.SetText("Center (0.7)")
	bot.SetText("Bottom-Left (0.3)")
	top.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).
		Background(tcell.ColorRed))
	mid.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).
		Background(tcell.ColorLime))
	bot.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlue))

	top.SetAlignment(views.VAlignTop | views.HAlignRight)
	mid.SetAlignment(views.VAlignCenter | views.HAlignCenter)
	bot.SetAlignment(views.VAlignBottom | views.HAlignLeft)

	box.SetOrientation(views.Vertical)
	box.AddWidget(title, 0)
	box.AddWidget(top, 0)
	box.AddWidget(mid, 0.7)
	box.AddWidget(bot, 0.3)

	app.SetRootWidget(box)
	if e := app.Run(); e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		os.Exit(1)
	}
}
