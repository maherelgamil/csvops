package csvops

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
)

func ptrStr(s string) *string    { return &s }
func ptrF64(f float64) *float64  { return &f }

func runFilter(t *testing.T, csvBody string, opts FilterOptions) (string, FilterResult) {
	t.Helper()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, csvBody)

	var buf bytes.Buffer
	opts.Input = in
	opts.Output = &buf
	if !opts.WithHeader {
		// keep behavior consistent with CLI default
	}
	res, err := Filter(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	return buf.String(), res
}

func TestFilter_Eq(t *testing.T) {
	got, res := runFilter(t, "id,country\n1,Egypt\n2,USA\n3,Egypt\n", FilterOptions{
		Column:     "country",
		Eq:         ptrStr("Egypt"),
		WithHeader: true,
	})
	if res.Matched != 2 {
		t.Errorf("Matched = %d, want 2", res.Matched)
	}
	want := "id,country\n1,Egypt\n3,Egypt\n"
	if got != want {
		t.Errorf("output:\n%s\nwant:\n%s", got, want)
	}
}

func TestFilter_EqEmptyString(t *testing.T) {
	got, res := runFilter(t, "id,note\n1,\n2,hi\n3,\n", FilterOptions{
		Column:     "note",
		Eq:         ptrStr(""),
		WithHeader: true,
	})
	if res.Matched != 2 {
		t.Errorf("Matched = %d, want 2 (empty values)", res.Matched)
	}
	if got != "id,note\n1,\n3,\n" {
		t.Errorf("got:\n%s", got)
	}
}

func TestFilter_AndSemantics(t *testing.T) {
	_, res := runFilter(t, "id,score\n1,50\n2,90\n3,95\n", FilterOptions{
		Column: "score",
		Gt:     ptrF64(80),
		Lt:     ptrF64(100),
		All:    true,
	})
	if res.Matched != 2 {
		t.Errorf("AND: Matched = %d, want 2", res.Matched)
	}
}

func TestFilter_OrSemantics(t *testing.T) {
	_, res := runFilter(t, "id,score\n1,50\n2,90\n3,200\n", FilterOptions{
		Column: "score",
		Lt:     ptrF64(60),
		Gt:     ptrF64(150),
	})
	if res.Matched != 2 {
		t.Errorf("OR: Matched = %d, want 2", res.Matched)
	}
}

func TestFilter_ContainsCaseInsensitive(t *testing.T) {
	_, res := runFilter(t, "name\nAlice\nbob\nALison\n", FilterOptions{
		Column:   "name",
		Contains: ptrStr("ali"),
	})
	if res.Matched != 2 {
		t.Errorf("Matched = %d, want 2", res.Matched)
	}
}

func TestFilter_NoConditionsErrors(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "a\n1\n")
	var buf bytes.Buffer
	_, err := Filter(context.Background(), FilterOptions{
		Input:  in,
		Output: &buf,
		Column: "a",
	})
	if err == nil {
		t.Error("expected error when no conditions set")
	}
}

func TestFilter_UnknownColumnErrors(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "a\n1\n")
	var buf bytes.Buffer
	_, err := Filter(context.Background(), FilterOptions{
		Input:  in,
		Output: &buf,
		Column: "missing",
		Eq:     ptrStr("x"),
	})
	if err == nil {
		t.Error("expected error for unknown column")
	}
}
