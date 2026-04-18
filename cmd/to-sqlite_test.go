package cmd

import "testing"

func TestQuoteIdent(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"users", `"users"`},
		{"my table", `"my table"`},
		{`weird"; drop table x; --`, `"weird""; drop table x; --"`},
		{`a"b"c`, `"a""b""c"`},
		{"", `""`},
	}
	for _, tt := range tests {
		got := quoteIdent(tt.in)
		if got != tt.want {
			t.Errorf("quoteIdent(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
