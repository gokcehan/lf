package views

import "testing"

func TestText(t *testing.T) {
	text := &Text{}

	text.SetText(`
This
String
Is
Pretty
Long
12345678901234567890
`)
	if text.width != 20 {
		t.Errorf("Incorrect width: %d, expected: %d", text.width, 20)
	}
}
