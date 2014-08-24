package duplo

import (
	"github.com/nfnt/resize"
	"github.com/rivo/duplo/haar"
	"image"
	"math"
	"math/rand"
)

// Hash represents the visual hash of an image.
type Hash struct {
	haar.Matrix

	// Thresholds contains the coefficient threholds. If you discard all
	// coefficients with abs(coef) < threshold, you end up with TopCoefs
	// coefficients.
	Thresholds haar.Coef

	// Ratio is image width / image height or 0 if height is 0.
	Ratio float64
}

// CreateHash calculates and returns the visual hash of the provided image.
func CreateHash(img image.Image) Hash {
	// Determine image ratio.
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	var ratio float64
	if height > 0 {
		ratio = float64(width) / float64(height)
	}

	// Resize the image.
	scaled := resize.Resize(ImageScale, ImageScale, img, resize.Bicubic)

	// Then perform a 2D Haar Wavelet transform.
	matrix := haar.Transform(scaled)

	// Find the kth largest coefficients for each colour channel.
	thresholds := coefThresholds(matrix.Coefs, TopCoefs)

	return Hash{haar.Matrix{matrix.Coefs, ImageScale, ImageScale}, thresholds, ratio}
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
