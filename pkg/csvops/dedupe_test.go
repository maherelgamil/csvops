package csvops

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestDedupe_KeepFirstPreservesOrder(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	out := filepath.Join(dir, "out.csv")
	writeCSV(t, in, "id,email\n1,a@x\n2,b@x\n3,a@x\n4,c@x\n5,b@x\n")

	res, err := Dedupe(context.Background(), DedupeOptions{
		Input:      in,
		Output:     out,
		KeyColumns: []string{"email"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.UniqueRows != 3 || res.Duplicates != 2 {
		t.Errorf("got unique=%d dup=%d, want 3/2", res.UniqueRows, res.Duplicates)
	}

	got := readFile(t, out)
	want := "id,email\n1,a@x\n2,b@x\n4,c@x\n"
	if got != want {
		t.Errorf("output:\n%s\nwant:\n%s", got, want)
	}
}

func TestDedupe_KeepLastPreservesOrder(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	out := filepath.Join(dir, "out.csv")
	writeCSV(t, in, "id,email\n1,a@x\n2,b@x\n3,a@x\n4,c@x\n5,b@x\n")

	res, err := Dedupe(context.Background(), DedupeOptions{
		Input:      in,
		Output:     out,
		KeyColumns: []string{"email"},
		KeepLast:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.UniqueRows != 3 {
		t.Errorf("UniqueRows = %d, want 3", res.UniqueRows)
	}

	got := readFile(t, out)
	// kept: 3 (a@x), 4 (c@x), 5 (b@x), in original-file order
	want := "id,email\n3,a@x\n4,c@x\n5,b@x\n"
	if got != want {
		t.Errorf("output:\n%s\nwant:\n%s", got, want)
	}
}

func TestDedupe_DeterministicAcrossRuns(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	body := "id,email\n"
	for i := 1; i <= 50; i++ {
		body += "1,dup@x\n"
	}
	writeCSV(t, in, body)

	first := ""
	for i := 0; i < 5; i++ {
		out := filepath.Join(dir, "out.csv")
		_, err := Dedupe(context.Background(), DedupeOptions{
			Input: in, Output: out, KeyColumns: []string{"email"},
		})
		if err != nil {
			t.Fatal(err)
		}
		got := readFile(t, out)
		if i == 0 {
			first = got
			continue
		}
		if got != first {
			t.Fatalf("non-deterministic output between runs:\n%s\nvs\n%s", first, got)
		}
	}
}

func TestDedupe_CompositeKey(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	out := filepath.Join(dir, "out.csv")
	writeCSV(t, in, "first,last,email\nAlice,Smith,a@x\nAlice,Jones,a@x\nAlice,Smith,b@x\n")

	res, err := Dedupe(context.Background(), DedupeOptions{
		Input:      in,
		Output:     out,
		KeyColumns: []string{"first", "last"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.UniqueRows != 2 {
		t.Errorf("UniqueRows = %d, want 2", res.UniqueRows)
	}
}

func TestDedupe_CaseSensitivity(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "name\nAlice\nalice\nALICE\n")

	out := filepath.Join(dir, "ci.csv")
	res, err := Dedupe(context.Background(), DedupeOptions{
		Input: in, Output: out, KeyColumns: []string{"name"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.UniqueRows != 1 {
		t.Errorf("case-insensitive: UniqueRows = %d, want 1", res.UniqueRows)
	}

	out2 := filepath.Join(dir, "cs.csv")
	res2, err := Dedupe(context.Background(), DedupeOptions{
		Input: in, Output: out2, KeyColumns: []string{"name"}, CaseSensitive: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res2.UniqueRows != 3 {
		t.Errorf("case-sensitive: UniqueRows = %d, want 3", res2.UniqueRows)
	}
}

func TestDedupe_InPlaceOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	writeCSV(t, path, "id,email\n1,a@x\n2,a@x\n3,b@x\n")

	_, err := Dedupe(context.Background(), DedupeOptions{
		Input: path, Output: path, KeyColumns: []string{"email"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	if !strings.Contains(got, "1,a@x") || !strings.Contains(got, "3,b@x") {
		t.Errorf("in-place output missing rows:\n%s", got)
	}
}

func TestDedupe_UnknownKey(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "a,b\n1,2\n")

	_, err := Dedupe(context.Background(), DedupeOptions{
		Input: in, Output: filepath.Join(dir, "out.csv"), KeyColumns: []string{"missing"},
	})
	if err == nil {
		t.Error("expected error for missing key column")
	}
}
