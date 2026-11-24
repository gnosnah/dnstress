package dnstress

import (
	"testing"
	"time"
)

func TestTopPercentileEmpty(t *testing.T) {
	s := &rttTopPercentileStats{max: 100, tp: 0.95}
	if got := s.topPercentile(); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestTopPercentileFewerThanMax(t *testing.T) {
	s := &rttTopPercentileStats{max: 100, tp: 0.95}
	s.append([]time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	})
	// ceil(0.95 * 5) = 5 → index 4 → 50ms
	if got := s.topPercentile(); got != 50*time.Millisecond {
		t.Fatalf("expected 50ms, got %v", got)
	}
}

func TestTopPercentileTruncatesToMax(t *testing.T) {
	s := &rttTopPercentileStats{max: 3, tp: 1.0}
	s.append([]time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		100 * time.Millisecond,
	})
	// only first 3 samples used after truncate, TP100 → 30ms
	if got := s.topPercentile(); got != 30*time.Millisecond {
		t.Fatalf("expected 30ms, got %v", got)
	}
}

func TestSafeRateAndQPS(t *testing.T) {
	if safeRate(0, 0) != 0 {
		t.Fatal("safeRate(0,0) should be 0")
	}
	if safeQPS(0, 1) != 0 {
		t.Fatal("safeQPS with total 0 should be 0")
	}
	if got := safeQPS(100, 0.0001); got != 100000 {
		t.Fatalf("safeQPS short elapsed = %d want 100000", got)
	}
	if safeRate(1, 2) != 50 {
		t.Fatalf("safeRate(1,2) = %v", safeRate(1, 2))
	}
}
