package tui

import "testing"

func TestMergeStreamContent(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		incoming string
		next     string
		delta    string
	}{
		{
			name:     "pure delta",
			current:  "Hello",
			incoming: " world",
			next:     "Hello world",
			delta:    " world",
		},
		{
			name:     "cumulative snapshot",
			current:  "Hello",
			incoming: "Hello world",
			next:     "Hello world",
			delta:    " world",
		},
		{
			name:     "stale snapshot",
			current:  "Hello world",
			incoming: "Hello",
			next:     "Hello world",
			delta:    "",
		},
		{
			name:     "overlap",
			current:  "Hello wor",
			incoming: "world",
			next:     "Hello world",
			delta:    "ld",
		},
		{
			name:     "duplicate resend",
			current:  "Hello",
			incoming: "Hello",
			next:     "Hello",
			delta:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			next, delta := mergeStreamContent(tc.current, tc.incoming)
			if next != tc.next || delta != tc.delta {
				t.Fatalf("got (%q, %q), want (%q, %q)", next, delta, tc.next, tc.delta)
			}
		})
	}
}
