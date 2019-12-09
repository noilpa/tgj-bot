package app

import "testing"

func TestIsMrTitleValid(t *testing.T) {
	tests := []struct {
		title string
		exp   bool
	}{
		{"foo [NC-123]bar", true},
		{"foo [NC-1230]bar", true},
		{"foo NC-1230 bar", false},
		{"foo [NC-1230 bar", false},
		{"foo [NC-1230 ]bar", false},
		{"foo NC-1230] bar", false},
		{"foo [ NC-1230] bar", false},
		{"foo [ NC-1230 ] bar", false},
		{"foo 1230 abr", true},
		{"foo 1230 NC abr", true},
	}

	for index, item := range tests {
		value := isMrTitleValid(item.title)
		if value != item.exp {
			t.Fatalf("failed at index %d", index)
		}
	}
}
