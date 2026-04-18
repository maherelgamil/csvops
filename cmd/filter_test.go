package cmd

import "testing"

func TestRowMatches(t *testing.T) {
	// rowMatches reads the package-level eqValue/containsValue/gtValue/ltValue.
	// Reset them around each subtest.
	defer func() {
		eqValue, containsValue = "", ""
		gtValue, ltValue = 0, 0
	}()

	t.Run("eq match", func(t *testing.T) {
		eqValue = "Egypt"
		if !rowMatches("Egypt", true, false, false, false, false) {
			t.Error("expected match")
		}
		if rowMatches("USA", true, false, false, false, false) {
			t.Error("unexpected match")
		}
	})

	t.Run("eq empty string matches when set", func(t *testing.T) {
		eqValue = ""
		if !rowMatches("", true, false, false, false, false) {
			t.Error("--eq=\"\" should match empty values")
		}
		if rowMatches("x", true, false, false, false, false) {
			t.Error("--eq=\"\" should not match non-empty")
		}
	})

	t.Run("contains case-insensitive", func(t *testing.T) {
		containsValue = "ali"
		if !rowMatches("Alice", false, true, false, false, false) {
			t.Error("contains should be case-insensitive")
		}
	})

	t.Run("gt ignores non-numeric", func(t *testing.T) {
		gtValue = 5
		if rowMatches("not-a-number", false, false, true, false, false) {
			t.Error("gt should not match non-numeric")
		}
		if !rowMatches("10", false, false, true, false, false) {
			t.Error("gt should match 10 > 5")
		}
	})

	t.Run("OR semantics", func(t *testing.T) {
		eqValue, containsValue = "Egypt", "xyz"
		if !rowMatches("Egypt", true, true, false, false, false) {
			t.Error("OR: should match if eq matches even if contains doesn't")
		}
	})

	t.Run("AND semantics", func(t *testing.T) {
		eqValue, containsValue = "Egypt", "xyz"
		if rowMatches("Egypt", true, true, false, false, true) {
			t.Error("AND: should not match if contains fails")
		}
		containsValue = "gyp"
		if !rowMatches("Egypt", true, true, false, false, true) {
			t.Error("AND: should match if both pass")
		}
	})

	t.Run("no flags set", func(t *testing.T) {
		if rowMatches("anything", false, false, false, false, false) {
			t.Error("with no flags set, should not match")
		}
	})
}
