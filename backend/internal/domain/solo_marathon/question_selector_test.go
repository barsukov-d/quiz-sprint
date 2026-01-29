package solo_marathon

import (
	"testing"
)

func TestSelectWeightedDifficulty(t *testing.T) {
	tests := []struct {
		name         string
		distribution map[string]float64
		samples      int
		wantRanges   map[string][2]float64 // min and max expected frequency
	}{
		{
			name: "Beginner level - mostly easy",
			distribution: map[string]float64{
				"easy":   0.8,
				"medium": 0.2,
				"hard":   0.0,
			},
			samples: 1000,
			wantRanges: map[string][2]float64{
				"easy":   {0.75, 0.85}, // expect ~80% ±5%
				"medium": {0.15, 0.25}, // expect ~20% ±5%
				"hard":   {0.0, 0.0},   // expect 0%
			},
		},
		{
			name: "Master level - mostly hard",
			distribution: map[string]float64{
				"easy":   0.0,
				"medium": 0.3,
				"hard":   0.7,
			},
			samples: 1000,
			wantRanges: map[string][2]float64{
				"easy":   {0.0, 0.0},    // expect 0%
				"medium": {0.25, 0.35},  // expect ~30% ±5%
				"hard":   {0.65, 0.75},  // expect ~70% ±5%
			},
		},
		{
			name: "Medium level - balanced",
			distribution: map[string]float64{
				"easy":   0.5,
				"medium": 0.4,
				"hard":   0.1,
			},
			samples: 1000,
			wantRanges: map[string][2]float64{
				"easy":   {0.45, 0.55}, // expect ~50% ±5%
				"medium": {0.35, 0.45}, // expect ~40% ±5%
				"hard":   {0.05, 0.15}, // expect ~10% ±5%
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Count occurrences
			counts := make(map[string]int)
			for i := 0; i < tt.samples; i++ {
				difficulty := selectWeightedDifficulty(tt.distribution)
				counts[difficulty]++
			}

			// Check frequencies
			for difficulty, expectedRange := range tt.wantRanges {
				frequency := float64(counts[difficulty]) / float64(tt.samples)
				minExpected := expectedRange[0]
				maxExpected := expectedRange[1]

				if frequency < minExpected || frequency > maxExpected {
					t.Errorf(
						"%s: got frequency %.2f, want between %.2f and %.2f (count: %d/%d)",
						difficulty,
						frequency,
						minExpected,
						maxExpected,
						counts[difficulty],
						tt.samples,
					)
				}
			}
		})
	}
}

func TestSelectWeightedDifficulty_EdgeCases(t *testing.T) {
	t.Run("empty distribution", func(t *testing.T) {
		distribution := map[string]float64{}
		result := selectWeightedDifficulty(distribution)

		// Should return fallback
		if result != "medium" {
			t.Errorf("expected fallback 'medium', got %s", result)
		}
	})

	t.Run("single option", func(t *testing.T) {
		distribution := map[string]float64{
			"hard": 1.0,
		}

		for i := 0; i < 10; i++ {
			result := selectWeightedDifficulty(distribution)
			if result != "hard" {
				t.Errorf("expected 'hard', got %s", result)
			}
		}
	})

	t.Run("zero weights filtered out", func(t *testing.T) {
		distribution := map[string]float64{
			"easy":   0.0,
			"medium": 0.0,
			"hard":   1.0,
		}

		for i := 0; i < 10; i++ {
			result := selectWeightedDifficulty(distribution)
			if result != "hard" {
				t.Errorf("expected 'hard', got %s", result)
			}
		}
	})
}
