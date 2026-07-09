package main

import (
	"errors"
	"testing"
)

// selections must never contain a newline path, saveFiles and writeSelection rely on this invariant
func TestToggleSelectionRefusesNewline(t *testing.T) {
	nav := &nav{selections: map[string]int{}}

	if err := nav.toggleSelection("/tmp/a\nb"); !errors.Is(err, errNewline) {
		t.Errorf("toggleSelection(newline path) = %v, want errNewline", err)
	}
	if len(nav.selections) != 0 {
		t.Error("newline path entered selections")
	}

	if err := nav.toggleSelection("/tmp/ok"); err != nil {
		t.Errorf("toggleSelection(clean path) = %v, want nil", err)
	}

	// unselecting a stale newline entry must still work so users can recover
	nav.selections["/tmp/x\ry"] = 5
	if err := nav.toggleSelection("/tmp/x\ry"); err != nil {
		t.Errorf("unselecting stale newline entry = %v, want nil", err)
	}
	if _, ok := nav.selections["/tmp/x\ry"]; ok {
		t.Error("stale newline entry not removed")
	}
}

// rename must refuse a newline target before touching the filesystem
func TestRenameRefusesNewlineTarget(t *testing.T) {
	nav := &nav{renameOldPath: "/tmp/a", renameNewPath: "/tmp/a\nb"}
	if err := nav.rename(); !errors.Is(err, errNewline) {
		t.Errorf("rename to newline target = %v, want errNewline", err)
	}
}
