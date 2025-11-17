package scanner

import "math"

func calculateShannonEntropy(s string) float64 {
	if s == "" {
		return 0
	}

	counts := make(map[rune]int)
	for _, r := range s {
		counts[r]++
	}

	var entropy float64
	length := float64(len(s))
	for _, count := range counts {
		// p(x) = frequency / total
		prob := float64(count) / length
		if prob > 0 {
			// H = -Σ p(x) * log2(p(x))
			entropy -= prob * math.Log2(prob)
		}
	}

	return entropy
}
