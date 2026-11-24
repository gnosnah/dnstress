package dnstress

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyRequiresStopCondition(t *testing.T) {
	cfg := defaultConfig()
	cfg.datafile = "example.txt"
	cfg.maxQueries = 0
	cfg.maxTime = 0
	if err := cfg.verify(); err == nil {
		t.Fatal("expected error when both maxQueries and maxTime are 0")
	}
	cfg.maxQueries = 10
	if err := cfg.verify(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidAddr(t *testing.T) {
	cases := map[string]bool{
		"127.0.0.1":  true,
		"::1":        true,
		"dns.google": true,
		"":           false,
		"bad host":   false,
	}
	for in, want := range cases {
		if got := validAddr(in); got != want {
			t.Fatalf("validAddr(%q)=%v want %v", in, got, want)
		}
	}
}

func TestVerifyNormalizesQTypeAndRejectsBadTimeout(t *testing.T) {
	cfg := defaultConfig()
	cfg.datafile = "example.txt"
	cfg.maxQueries = 1
	cfg.qType = "aaaa"
	if err := cfg.verify(); err != nil {
		t.Fatalf("unexpected error for lowercase qtype: %v", err)
	}
	if cfg.qType != "AAAA" {
		t.Fatalf("qtype not normalized: %q", cfg.qType)
	}

	cfg.timeout = 0
	if err := cfg.verify(); err == nil {
		t.Fatal("expected error for timeout <= 0")
	}
}

func TestLoadDataSkipsEmptyAndComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "q.txt")
	content := "# comment\n  # indented comment\n\n\ngoogle.com A\n\nexample.com\nbadtype XYZ\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewStress("127.0.0.1", 53, path, WithMaxQueries(1))
	qs, err := s.loadData(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 2 {
		t.Fatalf("expected 2 queries, got %d", len(qs))
	}
	if qs[0].dName != "google.com" || qs[0].dType != "A" {
		t.Fatalf("unexpected first item: %+v", qs[0])
	}
	if qs[1].dName != "example.com" || qs[1].dType != "" {
		t.Fatalf("unexpected second item: %+v", qs[1])
	}
}

func TestInitRejectsEmptyQueries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(path, []byte("# only comment\n\n"), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewStress("127.0.0.1", 53, path, WithMaxQueries(1))
	if err := s.Init(); err == nil {
		t.Fatal("expected error for empty query file")
	}
}
