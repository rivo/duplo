package duplo

// Match represents an image matched by a similarity query.
type Match struct {
	// The ID of the matched image, as specified in the pool.Add() function.
	ID interface{}

	// The score calculated during the similarity query. The lower, the better
	// the match.
	Score float64

	// The absolute difference between the two image ratios.
	RatioDiff float64

	// The hamming distance between the two cross-sections.
	HammingDistance int
}

type matches []*Match

func (m matches) Len() int           { return len(m) }
func (m matches) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m matches) Less(i, j int) bool { return m[i].Score < m[j].Score }
