package duplo

import (
	"fmt"
)

// Match represents an image matched by a similarity query.
type Match struct {
	// The ID of the matched image, as specified in the pool.Add() function.
	ID interface{}

	// The score calculated during the similarity query. The lower, the better
	// the match.
	Score float64

	// The absolute difference between the two image ratios' log values.
	RatioDiff float64

	// The hamming distance between the two dHash bit vectors.
	DHashDistance int

	// The hamming distance between the two histogram bit vectors.
	HistogramDistance int
}

// Matches is a slice of match results.
type Matches []*Match

func (m Matches) Len() int           { return len(m) }
func (m Matches) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m Matches) Less(i, j int) bool { return m[j] == nil || (m[i] != nil && m[i].Score < m[j].Score) }

func (m *Match) String() string {
	return fmt.Sprintf("%s: score=%.4f, ratio-diff=%.1f, dHash-dist=%d, histDist=%d",
		m.ID, m.Score, m.RatioDiff, m.DHashDistance, m.HistogramDistance)
}
