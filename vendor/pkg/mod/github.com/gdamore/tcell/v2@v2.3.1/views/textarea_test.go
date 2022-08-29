package views

import "testing"

func TestSetContent(t *testing.T) {
	ta := &TextArea{}

	ta.SetContent("This is a quite long line.")  // This line is longer than 11.
	ta.SetContent("Four.\nFive...\n...and Six.") //"...and Six." should be 11 long.

	if ta.model.height != 3 {
		t.Errorf("Incorrect height: %d, expected: %d", ta.model.height, 3)
	}
	if ta.model.width != 11 {
		t.Errorf("Incorrect width: %d, expected: %d", ta.model.width, 11)
	}
}
