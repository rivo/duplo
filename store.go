package duplo

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"math"
	"sync"

	"github.com/rivo/duplo/haar"
)

const (
	// ImageScale is the width and height to which images are resized before they
	// are being processed.
	ImageScale = 128
)

var (
	// TopCoefs is the number of top coefficients (per colour channel), ordered
	// by absolute value, that will be kept. Coefficients that rank lower will
	// be discarded. Change this only once when the package is initialized.
	TopCoefs = 40

	// The weights for the scoring function (currently for the YIQ colour space).
	weights = [3][6]float64{
		{5.00, 0.83, 1.01, 0.52, 0.47, 0.30},
		{19.21, 1.26, 0.44, 0.53, 0.28, 0.14},
		{34.37, 0.36, 0.45, 0.14, 0.18, 0.27},
	}

	// The weights, totalled over all colour channels.
	weightSums = [6]float64{58.58, 2.45, 1.9, 1.19, 0.93, 0.71}
)

// Store is a data structure that holds references to images. It holds visual
// hashes and references to the images but the images themselves are not held
// in the data structure.
//
// A general limit to the store is that it can hold no more than 4,294,967,295
// images. This is to save RAM space but may be easy to extend by modifying its
// data structures to hold uint64 indices instead of uint32 indices.
//
// Store's methods are concurrency safe. Store implements the GobDecoder and
// GobEncoder interfaces.
type Store struct {
	sync.RWMutex

	// All images in the store or, rather, the candidates for a query.
	candidates []candidate

	// All IDs in the store, mapping to candidate indices.
	ids map[interface{}]uint32

	// indices  contains references to the images in the store. It is a slice
	// of slices which contains image indices (into the "candidates" slice).
	// Use the following formula to access an index slice:
	//
	//		s := store.indices[sign*ImageScale*ImageScale*haar.ColourChannels + coefIdx*haar.ColourChannels + channel]
	//
	// where the variables are as follows:
	//
	//		* sign: Either 0 (positive) or 1 (negative)
	//		* coefIdx: The index of the coefficient (from 0 to (ImageScale*ImageScale)-1)
	//		* channel: The colour channel (from 0 to haar.ColourChannels-1)
	indices [][]uint32

	// Whether this store was modified since it was loaded/created.
	modified bool
}

// New returns a new, empty image store.
func New() *Store {
	store := new(Store)

	store.ids = make(map[interface{}]uint32)
	store.indices = make([][]uint32, 2*ImageScale*ImageScale*haar.ColourChannels)

	return store
}

// Has checks if an image (via its ID) is already contained in the store.
func (store *Store) Has(id interface{}) bool {
	store.RLock()
	defer store.RUnlock()

	_, ok := store.ids[id]
	return ok
}

// Add adds an image (via its hash) to the store. The provided ID is the value
// that will be returned as the result of a similarity query. If an ID is
// already in the store, it is not added again.
func (store *Store) Add(id interface{}, hash Hash) {
	store.Lock()
	defer store.Unlock()

	// Do we already manage this image?
	_, ok := store.ids[id]
	if ok {
		// Yes, we do. Don't add it again.
		return
	}

	// We need this for when we serialize the store.
	gob.Register(id)

	// Make this image a candidate.
	index := len(store.candidates)
	store.candidates = append(store.candidates, candidate{
		id,
		hash.Coefs[0],
		hash.Ratio,
		hash.DHash,
		hash.Histogram,
		hash.HistoMax})
	store.ids[id] = uint32(index)

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
			location := sign*ImageScale*ImageScale*haar.ColourChannels + coefIndex*haar.ColourChannels + colourIndex
			store.indices[location] = append(store.indices[location], uint32(index))
		}
	}

	// Image was successfully added.
	store.modified = true
}

// IDs returns a list of IDs of all images contained in the store. This list is
// created during the call so it may be modified without affecting the store.
func (store *Store) IDs() (ids []interface{}) {
	store.Lock()
	defer store.Unlock()

	for id := range store.ids {
		ids = append(ids, id)
	}

	return
}

// Delete removes an image from the store so it will not be returned during a
// query anymore. Note that the candidate slot still remains occupied but its
// index will be removed from all index lists. This also means that Size() will
// not decrease. This is an expensive operation. If the provided ID could not be
// found, nothing happens.
func (store *Store) Delete(id interface{}) {
	store.Lock()
	defer store.Unlock()

	// Get the index.
	index, ok := store.ids[id]
	if !ok {
		return // ID was not found.
	}
	store.modified = true

	// Clear the candidate.
	store.candidates[index].id = nil
	delete(store.ids, id)

	// Remove from all index lists.
	for location, list := range store.indices {
		for indexIndex := range list {
			if list[indexIndex] == index {
				store.indices[location] = append(list[:indexIndex], list[indexIndex+1:]...)
				break
			}
		}
	}
}

// Exchange exchanges the ID of an image for a new one. If the old ID could not
// be found, nothing happens. If the new ID already existed prior to the
// exchange, an error is returned.
func (store *Store) Exchange(oldID, newID interface{}) error {
	store.Lock()
	defer store.Unlock()

	// Get the old index.
	index, ok := store.ids[oldID]
	if !ok {
		return nil // ID was not found.
	}

	// Check if the new ID already exists.
	if _, ok := store.ids[newID]; ok {
		return fmt.Errorf("Cannot exchange ID, %s already exists", newID)
	}

	// Update the map.
	delete(store.ids, oldID)
	store.ids[newID] = index

	// Update the candidate.
	store.candidates[index].id = newID

	store.modified = true
	return nil
}

// Query performs a similarity search on the given image hash and returns
// all potential matches. The returned slice will not be sorted but implements
// sort.Interface, which will sort it so the match with the best score is its
// first element.
func (store *Store) Query(hash Hash) Matches {
	store.RLock()
	defer store.RUnlock()

	// Empty store, empty result set.
	if len(store.candidates) == 0 {
		return nil
	}

	// We're often touching all candidates at some point.
	scores := make([]float64, len(store.candidates))
	for index := range scores {
		scores[index] = math.NaN()
	}
	var numMatches int

	// Examine hash buckets.
	for coefIndex, coef := range hash.Coefs {
		if coefIndex == 0 {
			// Ignore scaling function coefficient for now.
			continue
		}

		// Calculate the weight bin outside the main loop.
		y := coefIndex / int(hash.Width)
		x := coefIndex % int(hash.Width)
		bin := y
		if x > y {
			bin = x
		}
		if bin > 5 {
			bin = 5
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

			location := sign*ImageScale*ImageScale*haar.ColourChannels + coefIndex*haar.ColourChannels + colourIndex
			for _, index := range store.indices[location] {
				// Do we know this index already?
				if math.IsNaN(scores[index]) {
					// No. Calculate initial score.
					score := 0.0
					for colour := range coef {
						score += weights[colour][0] *
							math.Abs(store.candidates[index].scaleCoef[colour]-hash.Coefs[0][colour])
					}
					scores[index] = score
				}

				// At this point, we have an entry in matches. Simply subtract the
				// corresponding weight.
				scores[index] -= weightSums[bin]
			}
		}
	}

	// Create matches.
	matches := make([]*Match, 0, numMatches)
	for index, score := range scores {
		if !math.IsNaN(score) {
			matches = append(matches, &Match{
				ID:        store.candidates[index].id,
				Score:     score,
				RatioDiff: math.Abs(math.Log(store.candidates[index].ratio) - math.Log(hash.Ratio)),
				DHashDistance: hammingDistance(store.candidates[index].dHash[0], hash.DHash[0]) +
					hammingDistance(store.candidates[index].dHash[1], hash.DHash[1]),
				HistogramDistance: hammingDistance(store.candidates[index].histogram, hash.Histogram),
			})
		}
	}

	return matches
}

// Size returns the number of images currently in the store.
func (store *Store) Size() int {
	store.RLock()
	defer store.RUnlock()

	return len(store.candidates)
}

// Modified indicates whether this store has been modified since it was loaded
// or created.
func (store *Store) Modified() bool {
	store.RLock()
	defer store.RUnlock()

	return store.modified
}

// GobDecode reconstructs the store from a binary representation. You may need
// to register any types that you put into the store in order for them to be
// decoded successfully. Example:
//
//     gob.Register(YourType{})
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
		if version < 2 {
			// Version 1 had a different coefficient type (slice instead of array).
			var coef []float64
			if err := decoder.Decode(&coef); err != nil {
				return fmt.Errorf("Unable to decode candidate scaling function coefficient: %s", err)
			}
			for i := range coef {
				store.candidates[index].scaleCoef[i] = coef[i]
			}
		} else {
			if err := decoder.Decode(&store.candidates[index].scaleCoef); err != nil {
				return fmt.Errorf("Unable to decode candidate scaling function coefficient: %s", err)
			}
		}
		if err := decoder.Decode(&store.candidates[index].ratio); err != nil {
			return fmt.Errorf("Unable to decode candidate ratio: %s", err)
		}
		if err := decoder.Decode(&store.candidates[index].dHash); err != nil {
			return fmt.Errorf("Unable to decode dHash: %s", err)
		}
		if err := decoder.Decode(&store.candidates[index].histogram); err != nil {
			return fmt.Errorf("Unable to decode histogram vector: %s", err)
		}
		if err := decoder.Decode(&store.candidates[index].histoMax); err != nil {
			return fmt.Errorf("Unable to decode histogram maximum: %s", err)
		}
	}

	// The ID set.
	if version < 3 {
		// Versions 1 and 2 used "int" indices. We need to convert.
		ids := make(map[interface{}]int)
		if err := decoder.Decode(&ids); err != nil {
			return fmt.Errorf("Unable to decode ID set: %s", err)
		}
		for key, value := range ids {
			store.ids[key] = uint32(value)
		}
	} else {
		if err := decoder.Decode(&store.ids); err != nil {
			return fmt.Errorf("Unable to decode ID set: %s", err)
		}
	}

	// The coefficient size.
	if version < 2 {
		// Version 1 had coefficient size in store.
		var coefSize int
		if err := decoder.Decode(&coefSize); err != nil {
			return fmt.Errorf("Unable to decode coefficient size: %s", err)
		}
	}

	// Indices.
	if version < 3 {
		// Versions 1 and 2 used "int" indices and a 4D matrix. We need to convert.
		var indices [][][][]int
		if err := decoder.Decode(&indices); err != nil {
			return fmt.Errorf("Unable to decode indices: %s", err)
		}
		for sign, s1 := range indices {
			for coefIndex, s2 := range s1 {
				for colourIndex, indexSlice := range s2 {
					location := sign*ImageScale*ImageScale*haar.ColourChannels + coefIndex*haar.ColourChannels + colourIndex
					store.indices[location] = make([]uint32, len(indexSlice))
					for i, index := range indexSlice {
						store.indices[location][i] = uint32(index)
					}
				}
			}
		}
		store.modified = true
	} else {
		if err := decoder.Decode(&store.indices); err != nil {
			return fmt.Errorf("Unable to decode indices: %s", err)
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
	if err := encoder.Encode(3); err != nil {
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
		if err := encoder.Encode(candidate.dHash); err != nil {
			return nil, fmt.Errorf("Unable to encode dHash: %s", err)
		}
		if err := encoder.Encode(candidate.histogram); err != nil {
			return nil, fmt.Errorf("Unable to encode histogram bit vector: %s", err)
		}
		if err := encoder.Encode(candidate.histoMax); err != nil {
			return nil, fmt.Errorf("Unable to encode histogram maximum: %s", err)
		}
	}

	// The ID set.
	if err := encoder.Encode(store.ids); err != nil {
		return nil, fmt.Errorf("Unable to encode ID set: %s", err)
	}

	// Indices.
	if err := encoder.Encode(store.indices); err != nil {
		return nil, fmt.Errorf("Unable to encode indices: %s", err)
	}

	// Finish up.
	compressor.Close()

	return buffer.Bytes(), nil
}
