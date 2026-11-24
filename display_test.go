package dnstress

import "testing"

func TestFormatElapsed(t *testing.T) {
	cases := map[float64]string{
		0:    "0.0s",
		3.2:  "3.2s",
		65.4: "1:05.4",
	}
	for in, want := range cases {
		if got := formatElapsed(in); got != want {
			t.Fatalf("formatElapsed(%v)=%q want %q", in, got, want)
		}
	}
}

func TestRenderBar(t *testing.T) {
	bar := renderBar(50, 10, false)
	if bar != "[█████░░░░░]" {
		t.Fatalf("unexpected bar: %q", bar)
	}
	bar = renderBar(150, 4, false)
	if bar != "[████]" {
		t.Fatalf("unexpected capped bar: %q", bar)
	}
}
