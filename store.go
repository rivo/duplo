package duplo

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"math"
	"sort"
	"sync"
)

var (
	// ImageScale is the width and height to which images are resized before they
	// are being processed. Change this only once when the package is initialized.
	ImageScale uint = 50

	// TopCoefs is the number of top coefficients (per colour channel), ordered
	// by absolute value, that will be kept. Coefficients that rank lower will
	// be discarded. Change this only once when the package is initialized.
	TopCoefs = 40

	// The weights for the scoring function (currently for the YIQ colour space
	// but we'll use it for YCbCr, too).
	weights = [3][6]float64{
		[6]float64{5.00, 0.83, 1.01, 0.52, 0.47, 0.30},
		[6]float64{19.21, 1.26, 0.44, 0.53, 0.28, 0.14},
		[6]float64{34.37, 0.36, 0.45, 0.14, 0.18, 0.27}}
)

// Store is a data structure that holds references to images. It holds visual
// hashes and references to the images but the images themselves are not held
// in the data structure.
//
// Store's methods are concurrency safe. Store implements the GobDecoder and
// GobEncoder interfaces.
type Store struct {
	sync.RWMutex

	// All images in the store or, rather, the candidates for a query.
	candidates []candidate

	// The number of elements (colour channels) in a coefficient.
	coefSize int

	// indices is a matrix which contains references to the images in the
	// store. At the tail of the matrix is an index into the candidates field.
	// The dimensions of this matrix are as follows: coefficient sign (0=positive,
	// 1=negative), coefficient index (from 0 to (ImageScale*ImageScale)-1),
	// colour space (from 0 to coefSize-1). All of these dimensions specified, one
	// will either find a nil slice (no images stored under that node) or a slice
	// of indices in the candidates field.
	indices [][][][]int

	// Whether this store was modified since it was loaded/created.
	Modified bool
}

// NewStore returns a new, empty image store.
func NewStore() *Store {
	store := new(Store)

	store.indices = make([][][][]int, 2)
	for index := range store.indices {
		store.indices[index] = make([][][]int, ImageScale*ImageScale)
	}

	return store
}

// Add adds an image (via its hash) to the store. The provided ID is the value
// that will be returned as the result of a similarity query.
func (store *Store) Add(id interface{}, hash Hash) {
	store.Lock()
	defer store.Unlock()

	// We need this for when we serialize the store.
	gob.Register(id)

	// We may not have enough space to add this image yet. If so, make some.
	if len(hash.Thresholds) > store.coefSize {
		for signIndex := range store.indices {
			for coefIndex := range store.indices[signIndex] {
				store.indices[signIndex][coefIndex] = append(store.indices[signIndex][coefIndex],
					make([][]int, len(hash.Thresholds)-store.coefSize)...)
			}
		}
	}
	store.coefSize = len(hash.Thresholds)

	// Make this image a candidate.
	index := len(store.candidates)
	store.candidates = append(store.candidates, candidate{
		id:           id,
		scaleCoef:    hash.Coefs[0],
		ratio:        hash.Ratio,
		crossSection: hash.CrossSection})

	// Distribute candidate index into the buckets.
	for coefIndex, coef := range hash.Coefs {
		if coefIndex == 0 {
			// This is the scaling function coefficient. Ignore.
			continue
		}

		for colourIndex, colourCoef := range coef {
			if math.Abs(colourCoef) < hash.Thresholds[colourIndex] {
				// Coef is too small. Ignore.
				continue
			}

			sign := 0
			if colourCoef < 0 {
				sign = 1
			}

			// Add this image's index to the bucket.
			store.indices[sign][coefIndex][colourIndex] =
				append(store.indices[sign][coefIndex][colourIndex], index)
		}
	}

	// Image was successfully added.
	store.Modified = true
}

// Query performs a similarity search on the given image hash and returns
// all potential matches, sorted by their score (best, i.e. lowest, first).
func (store *Store) Query(hash Hash) []*Match {
	store.RLock()
	defer store.RUnlock()

	// Empty store, empty result set.
	if len(store.candidates) == 0 {
		return nil
	}

	// matchMap contains all candidates which were found in index buckets
	// during the search.
	matchMap := make(map[int]*Match, TopCoefs)

	// Examine hash buckets.
	for coefIndex, coef := range hash.Coefs {
		if coefIndex == 0 {
			// Ignore scaling function coefficient for now.
			continue
		}

		for colourIndex, colourCoef := range coef {
			if math.Abs(colourCoef) < hash.Thresholds[colourIndex] {
				// Coef is too small. Ignore.
				continue
			}

			// At this point, we have a coefficient which we want to look up
			// in the index buckets.

			sign := 0
			if colourCoef < 0 {
				sign = 1
			}

			for _, index := range store.indices[sign][coefIndex][colourIndex] {
				// Do we know this index already?
				if _, ok := matchMap[index]; !ok {
					// No. Calculate initial score.
					score := 0.0
					for colour := range coef {
						score += weights[colour][0] *
							math.Abs(store.candidates[index].scaleCoef[colour]-hash.Coefs[0][colour])
					}

					// Difference in image ratios.
					ratioDiff := math.Abs(store.candidates[index].ratio - hash.Ratio)

					// The hamming distance between the cross-section bit vectors.
					hamming := hammingDistance(store.candidates[index].crossSection, hash.CrossSection)

					// Add this match.
					matchMap[index] = &Match{
						store.candidates[index].id,
						score,
						ratioDiff,
						hamming}
				}

				// At this point, we have an entry in matches. Simply subtract the
				// corresponding weight.
				for colour := range coef {
					bin := hash.Height
					if hash.Width > hash.Height {
						bin = hash.Width
					}
					if bin > 5 {
						bin = 5
					}
					matchMap[index].Score -= weights[colour][bin]
				}
			}
		}
	}

	// Make a sorted list of matches.
	result := make(matches, 0, len(matchMap))
	for _, match := range matchMap {
		result = append(result, match)
	}
	sort.Sort(matches(result))

	return result
}

// Size returns the number of images currently in the store.
func (store *Store) Size() int {
	store.RLock()
	defer store.RUnlock()

	return len(store.candidates)
}

// GobDecode reconstructs the store from a binary representation.
func (store *Store) GobDecode(from []byte) error {
	store.Lock()
	defer store.Unlock()

	buffer := bytes.NewReader(from)
	decompressor, err := gzip.NewReader(buffer)
	if err != nil {
		return fmt.Errorf("Unable to open decompressor: %s", err)
	}
	defer decompressor.Close()
	decoder := gob.NewDecoder(decompressor)

	// Do we have a version compatibility problem?
	var version int
	if err := decoder.Decode(&version); err != nil {
		return fmt.Errorf("Unable to decode store version: %s", err)
	}
	// So far, all previous versions accepted.

	// Candidates.
	var size int
	if err := decoder.Decode(&size); err != nil {
		return fmt.Errorf("Unable to decode candidate length: %s", err)
	}
	store.candidates = make([]candidate, size)
	for index := 0; index < size; index++ {
		if err := decoder.Decode(&store.candidates[index].id); err != nil {
			return fmt.Errorf("Unable to decode candidate ID: %s", err)
		}
		if err := decoder.Decode(&store.candidates[index].scaleCoef); err != nil {
			return fmt.Errorf("Unable to decode candidate scaling function coefficient: %s", err)
		}
		if err := decoder.Decode(&store.candidates[index].ratio); err != nil {
			return fmt.Errorf("Unable to decode candidate ratio: %s", err)
		}
	}

	// Indices.
	if err := decoder.Decode(&store.indices); err != nil {
		return fmt.Errorf("Unable to decode indices: %s", err)
	}

	// Complete the store.
	store.coefSize = 0
	for _, candidate := range store.candidates {
		if size := len(candidate.scaleCoef); size > store.coefSize {
			store.coefSize = size
		}
	}

	return nil
}

// GobEncode places a binary representation of the store in a byte slice.
func (store *Store) GobEncode() ([]byte, error) {
	store.RLock()
	defer store.RUnlock()

	buffer := new(bytes.Buffer)
	compressor := gzip.NewWriter(buffer)
	encoder := gob.NewEncoder(compressor)

	// Add a version number first.
	if err := encoder.Encode(1); err != nil {
		return nil, fmt.Errorf("Unable to encode store version: %s", err)
	}

	// Candidates are encoded manually because the encoder does not have access
	// to the candidate struct.
	if err := encoder.Encode(len(store.candidates)); err != nil {
		return nil, fmt.Errorf("Unable to encode candidate length: %s", err)
	}
	for _, candidate := range store.candidates {
		if err := encoder.Encode(&candidate.id); err != nil {
			return nil, fmt.Errorf("Unable to encode candidate ID: %s", err)
		}
		if err := encoder.Encode(candidate.scaleCoef); err != nil {
			return nil, fmt.Errorf("Unable to encode candidate scaling function coefficient: %s", err)
		}
		if err := encoder.Encode(candidate.ratio); err != nil {
			return nil, fmt.Errorf("Unable to encode candidate ratio: %s", err)
		}
	}

	// Indices.
	if err := encoder.Encode(store.indices); err != nil {
		return nil, fmt.Errorf("Unable to encode indices: %s", err)
	}

	// Finish up.
	compressor.Close()

	return buffer.Bytes(), nil
}
