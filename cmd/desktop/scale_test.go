package main

import "testing"

func TestComputeScale(t *testing.T) {
	cases := []struct {
		width int
		want  float64
	}{
		{1280, 1.0},  // Steam Deck native
		{1600, 1.25}, // 1.25 raw
		{1920, 1.5},  // FHD
		{2560, 2.0},  // QHD
		{3840, 2.0},  // 4K → clamped to cap
		{800, 1.0},   // below native → clamped to floor
		{1440, 1.25}, // 1.125 → rounds to 1.25
		{1520, 1.25}, // 1.1875 → (1.1875*4=4.75→round 5→1.25)
	}
	for _, c := range cases {
		if got := computeScale(c.width); got != c.want {
			t.Errorf("computeScale(%d) = %v, want %v", c.width, got, c.want)
		}
	}
}
