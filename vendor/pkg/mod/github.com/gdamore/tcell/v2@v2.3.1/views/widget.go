// Copyright 2015 The Tcell Authors
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

package views

import (
	"github.com/gdamore/tcell/v2"
)

// Widget is the base object that all onscreen elements implement.
type Widget interface {
	// Draw is called to inform the widget to draw itself.  A containing
	// Widget will generally call this during the application draw loop.
	Draw()

	// Resize is called in response to a resize of the View.  Unlike with
	// other events, Resize performed by parents first, and they must
	// then call their children.  This is because the children need to
	// see the updated sizes from the parents before they are called.
	// In general this is done *after* the views have updated.
	Resize()

	// HandleEvent is called to ask the widget to handle any events.
	// If the widget has consumed the event, it should return true.
	// Generally, events are handled by the lower layers first, that
	// is for example, a button may have a chance to handle an event
	// before the enclosing window or panel.
	//
	// Its expected that Resize events are consumed by the outermost
	// Widget, and the turned into a Resize() call.
	HandleEvent(ev tcell.Event) bool

	// SetView is used by callers to set the visual context of the
	// Widget.  The Widget should use the View as a context for
	// drawing.
	SetView(view View)

	// Size returns the size of the widget (content size) as width, height
	// in columns.  Layout managers should attempt to ensure that at least
	// this much space is made available to the View for this Widget.  Extra
	// space may be allocated on as an needed basis.
	Size() (int, int)

	// Watch is used to register an interest in this widget's events.
	// The handler will receive EventWidget events for this widget.
	// The order of event delivery when there are multiple watchers is
	// not specified, and may change from one event to the next.
	Watch(handler tcell.EventHandler)

	// Unwatch is used to urnegister an interest in this widget's events.
	Unwatch(handler tcell.EventHandler)
}

// EventWidget is an event delivered by a specific widget.
type EventWidget interface {
	Widget() Widget
	tcell.Event
}

type widgetEvent struct {
	widget Widget
	tcell.EventTime
}

func (wev *widgetEvent) Widget() Widget {
	return wev.widget
}

func (wev *widgetEvent) SetWidget(widget Widget) {
	wev.widget = widget
}

// WidgetWatchers provides a common implementation for base Widget
// Watch and Unwatch interfaces, suitable for embedding in more concrete
// widget implementations.
type WidgetWatchers struct {
	watchers map[tcell.EventHandler]struct{}
}

// Watch monitors this WidgetWatcher, causing the handler to be fired
// with EventWidget as they are occur on the watched Widget.
func (ww *WidgetWatchers) Watch(handler tcell.EventHandler) {
	if ww.watchers == nil {
		ww.watchers = make(map[tcell.EventHandler]struct{})
	}
	ww.watchers[handler] = struct{}{}
}

// Unwatch stops monitoring this WidgetWatcher. The handler will no longer
// be fired for Widget events.
func (ww *WidgetWatchers) Unwatch(handler tcell.EventHandler) {
	if ww.watchers != nil {
		delete(ww.watchers, handler)
	}
}

// PostEvent delivers the EventWidget to all registered watchers.  It is
// to be called by the Widget implementation.
func (ww *WidgetWatchers) PostEvent(wev EventWidget) {
	for watcher := range ww.watchers {
		// Deliver events to all listeners, ignoring return value.
		watcher.HandleEvent(wev)
	}
}

// PostEventWidgetContent is called by the Widget when its content is
// changed, delivering EventWidgetContent to all watchers.
func (ww *WidgetWatchers) PostEventWidgetContent(w Widget) {
	ev := &EventWidgetContent{}
	ev.SetWidget(w)
	ev.SetEventNow()
	ww.PostEvent(ev)
}

// PostEventWidgetResize is called by the Widget when the underlying View
// has resized, delivering EventWidgetResize to all watchers.
func (ww *WidgetWatchers) PostEventWidgetResize(w Widget) {
	ev := &EventWidgetResize{}
	ev.SetWidget(w)
	ev.SetEventNow()
	ww.PostEvent(ev)
}

// PostEventWidgetMove is called by the Widget when it is moved to a new
// location, delivering EventWidgetMove to all watchers.
func (ww *WidgetWatchers) PostEventWidgetMove(w Widget) {
	ev := &EventWidgetMove{}
	ev.SetWidget(w)
	ev.SetEventNow()
	ww.PostEvent(ev)
}

// XXX: WidgetExposed, Hidden?
// XXX: WidgetExposed, Hidden?

// EventWidgetContent is fired whenever a widget's content changes.
type EventWidgetContent struct {
	widgetEvent
}

// EventWidgetResize is fired whenever a widget is resized.
type EventWidgetResize struct {
	widgetEvent
}

// EventWidgetMove is fired whenver a widget changes location.
type EventWidgetMove struct {
	widgetEvent
}
