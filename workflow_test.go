package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestTable(t *testing.T) {
	got, err := table(map[string]any{"id": 42, "name": "Mochi"}, []string{"id", "name"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "┌────┬───────┐\n│ ID │ NAME  │\n├────┼───────┤\n│ 42 │ Mochi │\n└────┴───────┘\n"; string(got) != want {
		t.Fatalf("table = %q, want %q", got, want)
	}
}

func TestGoFilesStayWithin100Lines(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		lines := bytes.Count(data, []byte{'\n'})
		if len(data) > 0 && data[len(data)-1] != '\n' {
			lines++
		}
		if lines > 100 {
			t.Errorf("%s has %d lines", file, lines)
		}
	}
}
