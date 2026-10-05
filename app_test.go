package main

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v3"
)

type mockScreen struct {
	tcell.Screen
	eventQ chan tcell.Event
}

func (m *mockScreen) Init() error                                { return nil }
func (m *mockScreen) Fini()                                      {}
func (m *mockScreen) Size() (int, int)                           { return 80, 24 }
func (m *mockScreen) EventQ() chan tcell.Event                   { return m.eventQ }
func (m *mockScreen) PutStrStyled(int, int, string, tcell.Style) {}
func (m *mockScreen) Show()                                      {}
func (m *mockScreen) Clear()                                     {}
func (m *mockScreen) HideCursor()                                {}
func (m *mockScreen) ShowCursor(int, int)                        {}
func (m *mockScreen) LockRegion(int, int, int, int, bool)        {}
func (m *mockScreen) Suspend() error                             { return nil }
func (m *mockScreen) Resume() error                              { return nil }

func setupTestApp(t *testing.T) *app {
	t.Helper()
	gSingleMode = true

	screen := &mockScreen{
		eventQ: make(chan tcell.Event, 100),
	}

	ui := newUI(screen)
	nav := newNav(ui)
	app := newApp(ui, nav)

	return app
}

func TestCqCommand(t *testing.T) {
	tests := []struct {
		name       string
		cmdName    string
		args       []string
		wantExit   bool
		wantCode   int
		wantForce  bool
		wantErrSub string
	}{
		{
			name:      "cq default exit code 1",
			cmdName:   "cq",
			args:      nil,
			wantExit:  true,
			wantCode:  1,
			wantForce: true,
		},
		{
			name:      "cquit alias default exit code 1",
			cmdName:   "cquit",
			args:      nil,
			wantExit:  true,
			wantCode:  1,
			wantForce: true,
		},
		{
			name:      "cq custom exit code 2",
			cmdName:   "cq",
			args:      []string{"2"},
			wantExit:  true,
			wantCode:  2,
			wantForce: true,
		},
		{
			name:      "cq exit code 0",
			cmdName:   "cq",
			args:      []string{"0"},
			wantExit:  true,
			wantCode:  0,
			wantForce: true,
		},
		{
			name:      "cquit custom exit code 42",
			cmdName:   "cquit",
			args:      []string{"42"},
			wantExit:  true,
			wantCode:  42,
			wantForce: true,
		},
		{
			name:       "cq invalid non-numeric argument",
			cmdName:    "cq",
			args:       []string{"abc"},
			wantExit:   false,
			wantErrSub: "invalid syntax",
		},
		{
			name:       "cq too many arguments",
			cmdName:    "cq",
			args:       []string{"1", "2"},
			wantExit:   false,
			wantErrSub: "too many arguments",
		},
		{
			name:       "cquit too many arguments",
			cmdName:    "cquit",
			args:       []string{"1", "2", "3"},
			wantExit:   false,
			wantErrSub: "too many arguments",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := setupTestApp(t)

			call := &callExpr{name: tc.cmdName, args: tc.args, count: 1}
			call.eval(app, nil)

			if tc.wantExit {
				select {
				case q := <-app.quitChan:
					if q.code != tc.wantCode {
						t.Errorf("got exit code %d, want %d", q.code, tc.wantCode)
					}
					if q.force != tc.wantForce {
						t.Errorf("got force %v, want %v", q.force, tc.wantForce)
					}
				default:
					t.Fatalf("expected quitChan to receive a message, but it was empty")
				}
			} else {
				select {
				case q := <-app.quitChan:
					t.Fatalf("expected no quit message, but received %+v", q)
				default:
				}
				if tc.wantErrSub != "" && !strings.Contains(app.ui.msg, tc.wantErrSub) {
					t.Errorf("expected ui error message containing %q, got %q", tc.wantErrSub, app.ui.msg)
				}
			}
		})
	}
}

func TestCqLoopImmediateExitWithJobs(t *testing.T) {
	tests := []struct {
		name       string
		setJobs    func(*app)
		callCmd    func(*app)
		wantExit   bool
		wantCode   int
		wantErrMsg string
	}{
		{
			name: "quit blocked when copyJobs > 0",
			setJobs: func(a *app) {
				a.nav.copyJobs = 2
			},
			callCmd: func(a *app) {
				a.requestQuit()
			},
			wantExit:   false,
			wantErrMsg: "quit: copy operation in progress",
		},
		{
			name: "quit blocked when moveTotal > 0",
			setJobs: func(a *app) {
				a.nav.moveTotal = 5
			},
			callCmd: func(a *app) {
				a.requestQuit()
			},
			wantExit:   false,
			wantErrMsg: "quit: move operation in progress",
		},
		{
			name: "quit blocked when deleteTotal > 0",
			setJobs: func(a *app) {
				a.nav.deleteTotal = 3
			},
			callCmd: func(a *app) {
				a.requestQuit()
			},
			wantExit:   false,
			wantErrMsg: "quit: delete operation in progress",
		},
		{
			name: "cq exits immediately despite copyJobs > 0",
			setJobs: func(a *app) {
				a.nav.copyJobs = 2
			},
			callCmd: func(a *app) {
				a.requestCq(1)
			},
			wantExit: true,
			wantCode: 1,
		},
		{
			name: "cq exits immediately despite moveTotal > 0",
			setJobs: func(a *app) {
				a.nav.moveTotal = 5
			},
			callCmd: func(a *app) {
				a.requestCq(2)
			},
			wantExit: true,
			wantCode: 2,
		},
		{
			name: "cq exits immediately despite deleteTotal > 0",
			setJobs: func(a *app) {
				a.nav.deleteTotal = 3
			},
			callCmd: func(a *app) {
				a.requestCq(7)
			},
			wantExit: true,
			wantCode: 7,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := setupTestApp(t)

			tc.setJobs(app)

			done := make(chan int, 1)
			go func() {
				code := app.loop()
				done <- code
			}()

			tc.callCmd(app)

			if tc.wantExit {
				select {
				case code := <-done:
					if code != tc.wantCode {
						t.Errorf("got loop exit code %d, want %d", code, tc.wantCode)
					}
				case <-time.After(2 * time.Second):
					t.Fatalf("timed out waiting for app.loop() to exit")
				}
			} else {
				time.Sleep(50 * time.Millisecond)
				select {
				case code := <-done:
					t.Fatalf("expected loop to stay running, but it exited with %d", code)
				default:
				}
				if tc.wantErrMsg != "" && !strings.Contains(app.ui.msg, tc.wantErrMsg) {
					t.Errorf("expected ui error message containing %q, got %q", tc.wantErrMsg, app.ui.msg)
				}
				app.requestCq(0)
				select {
				case <-done:
				case <-time.After(2 * time.Second):
					t.Fatalf("timed out terminating test loop")
				}
			}
		})
	}
}

func TestCqCompletion(t *testing.T) {
	matches, _ := matchCmd("cq")
	foundCq := false
	for _, m := range matches {
		if m.name == "cq" {
			foundCq = true
			break
		}
	}
	if !foundCq {
		t.Errorf("expected 'cq' in matchCmd matches, got: %v", matches)
	}

	matchesQuit, _ := matchCmd("cquit")
	foundCquit := false
	for _, m := range matchesQuit {
		if m.name == "cquit" {
			foundCquit = true
			break
		}
	}
	if !foundCquit {
		t.Errorf("expected 'cquit' in matchCmd matches, got: %v", matchesQuit)
	}
}
