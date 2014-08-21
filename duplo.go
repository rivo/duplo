package duplo

import (
	"bytes"
	"encoding/gob"
	"github.com/nfnt/resize"
	"github.com/rivo/duplo/haar"
	"image"
	"math"
	"math/rand"
)

var (
	// ImageScale is the width and height to which images are resized before they
	// are being processed.
	ImageScale uint = 50

	// TopCoefs is the number of top coefficients (per colour channel), ordered
	// by absolute value, that will be kept. Coefficients that rank lower will
	// be discarded.
	TopCoefs = 40
)

// Hash represents the visual hash of an image.
type Hash struct {
	// Coefs is the coefficients matrix as calculated by the 2D Haar Wavelet
	// transform. Its size is ImageScale*ImageScale and each element has as many
	// values as colour channels in the original image.
	Coefs haar.Matrix

	// Thresholds contains the coefficient threholds. If you discard all
	// coefficients with abs(coef) < threshold, you end up with TopCoefs
	// coefficients.
	Thresholds haar.Coef
}

// CreateHash calculates and returns the visual hash of the provided image.
func CreateHash(img image.Image) Hash {
	// Resize the image first.
	scaled := resize.Resize(ImageScale, ImageScale, img, resize.Bicubic)

	// Then perform a 2D Haar Wavelet transform.
	matrix := haar.Transform(scaled)

	// Find the kth largest coefficients for each colour channel.
	thresholds := coefThresholds(matrix.Coefs, TopCoefs)

	return Hash{matrix, thresholds}
}

// Store is a data structure that holds references to images. It holds visual
// hashes and references to the images but the images themselves are not held
// in the data structure.
//
// Store implements the GobDecoder and GobEncoder interfaces.
type Store struct {
	Size     int  // The number of images currently contained in the store.
	Modified bool // Whether this store was modified since it was loaded/created.
}

// NewStore returns a new, empty image store.
func NewStore() *Store {
	return new(Store)
}

// Add adds an image (via its hash) to the store. The provided ID is the value
// that will be returned as the result of a similarity query.
func (store *Store) Add(id interface{}, hash Hash) error {
	// Image was successfully added.
	store.Size++
	store.Modified = true
	return nil
}

// GobDecode reconstructs the store from a binary representation.
func (store *Store) GobDecode(from []byte) error {
	buffer := bytes.NewBuffer(from)
	decoder := gob.NewDecoder(buffer)

	// Do we have a version compatibility problem?
	var version int
	if err := decoder.Decode(&version); err != nil {
		return err
	}
	// So far, all previous versions accepted.

	// Get the size.
	if err := decoder.Decode(&store.Size); err != nil {
		return err
	}

	return nil
}

// GobEncode places a binary representation of the store in a byte slice.
func (store *Store) GobEncode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)

	// Add a version number first.
	if err := encoder.Encode(1); err != nil {
		return nil, err
	}

	// We want to know the size without going through the structure.
	if err := encoder.Encode(store.Size); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// coefThreshold returns, for the given coefficients, the kth largest absolute
// value. Only the nth element in each Coef is considered. If you discard all
// values v with abs(v) < threshold, you will end up with k values.
func coefThreshold(coefs []haar.Coef, k int, n int) float64 {
	// It's the QuickSelect algorithm.
	randomIndex := rand.Intn(len(coefs))
	pivot := math.Abs(coefs[randomIndex][n])
	leftCoefs := make([]haar.Coef, 0, len(coefs))
	rightCoefs := make([]haar.Coef, 0, len(coefs))

	for _, coef := range coefs {
		if math.Abs(coef[n]) > pivot {
			leftCoefs = append(leftCoefs, coef)
		} else if math.Abs(coef[n]) < pivot {
			rightCoefs = append(rightCoefs, coef)
		}
	}

	if k <= len(leftCoefs) {
		return coefThreshold(leftCoefs, k, n)
	} else if k > len(coefs)-len(rightCoefs) {
		return coefThreshold(rightCoefs, k-(len(coefs)-len(rightCoefs)), n)
	} else {
		return pivot
	}
}

// coefThreshold returns, for the given coefficients, the kth largest absolute
// values per colour channel. If you discard all values v with
// abs(v) < threshold, you will end up with k values.
func coefThresholds(coefs []haar.Coef, k int) haar.Coef {
	// No data, no thresholds.
	if len(coefs) == 0 {
		return haar.Coef{}
	}

	// Select thresholds.
	thresholds := make(haar.Coef, len(coefs[0]))
	for index := range thresholds {
		thresholds[index] = coefThreshold(coefs, k, index)
	}

	return thresholds
}
