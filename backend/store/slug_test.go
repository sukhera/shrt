package store

import (
	"strings"
	"testing"
)

func TestRandomSlug_LengthAndAlphabet(t *testing.T) {
	lengths := []int{1, 7, 16, 64}
	for _, n := range lengths {
		slug, err := randomSlug(n)
		if err != nil {
			t.Fatalf("randomSlug(%d): %v", n, err)
		}
		if len(slug) != n {
			t.Errorf("length: got %d, want %d", len(slug), n)
		}
		for _, c := range slug {
			if !strings.ContainsRune(base62Alphabet, c) {
				t.Errorf("slug %q contains non-base62 char %q", slug, c)
			}
		}
	}
}

func TestRandomSlug_Uniqueness(t *testing.T) {
	// Not a statistical guarantee, but a 7-char base62 space (62^7 ≈ 3.5e12)
	// makes collisions across 10k draws vanishingly unlikely; any repeat signals
	// a broken generator (e.g. constant or low-entropy output).
	const draws = 10000
	seen := make(map[string]struct{}, draws)
	for i := 0; i < draws; i++ {
		slug, err := randomSlug(7)
		if err != nil {
			t.Fatalf("randomSlug: %v", err)
		}
		if _, dup := seen[slug]; dup {
			t.Fatalf("duplicate slug %q after %d draws", slug, i)
		}
		seen[slug] = struct{}{}
	}
}

func TestRandomSlug_DistributionNoBias(t *testing.T) {
	// Rejection sampling against 248 should keep the alphabet roughly uniform.
	// With 62*2000 = 124k symbols over 62 buckets, each bucket expects ~2000;
	// allow a wide tolerance to avoid flakiness while still catching gross bias
	// (e.g. modulo skew favouring the first symbols).
	counts := make(map[rune]int)
	const draws = 2000
	for i := 0; i < draws; i++ {
		slug, err := randomSlug(62)
		if err != nil {
			t.Fatalf("randomSlug: %v", err)
		}
		for _, c := range slug {
			counts[c]++
		}
	}
	if len(counts) != 62 {
		t.Errorf("expected all 62 symbols to appear, got %d distinct", len(counts))
	}
	for c, n := range counts {
		if n < 1000 || n > 3000 {
			t.Errorf("symbol %q count %d outside tolerance [1000,3000]", c, n)
		}
	}
}
