package duplo

import (
	"github.com/nfnt/resize"
	"github.com/rivo/duplo/haar"
	"image"
	"image/color"
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

	// CrossSection is a 64 bit vector which represents a diagonal cross-section
	// of the image, 32 bits for the Y colour space and 16 bits each for the Cb
	// and Cr colour spaces. A bit is set when its pixel value is higher or equal
	// than the average pixel value.
	CrossSection uint64
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

	// Resize the image for the Wavelet transform.
	scaled := resize.Resize(ImageScale, ImageScale, img, resize.Bicubic)

	// Then perform a 2D Haar Wavelet transform.
	matrix := haar.Transform(scaled)

	// Find the kth largest coefficients for each colour channel.
	thresholds := coefThresholds(matrix.Coefs, TopCoefs)

	// Create the cross-section value.
	bits := crossSection(img)

	return Hash{haar.Matrix{matrix.Coefs, ImageScale, ImageScale}, thresholds, ratio, bits}
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

// Computes a 64 bit vector by sampling a 32 pixel diagonal cross-section of
// the image and setting the corresponding bit to 1 if it is higher or equal
// to the average pixel value and to 0 if it is lower than the average pixel
// value. We use 32 bits for the Y values and 16 bits for the Cb and Cr values
// each.
func crossSection(img image.Image) (bits uint64) {
	// Resize the image so we have 32 pixels in each diagonal.
	// We use 32 bits for Y, and 16 bits each for Cb and Cr. 64 in total.
	scaled := resize.Resize(32, 32, img, resize.Bicubic)

	// Read out diagonal.
	pixels := new([3][32]int)
	averages := new([3]int)
	var (
		y, cb, cr      int
		prevCb, prevCr int
	)
	for pos := 0; pos < 32; pos++ {
		colour := scaled.At(pos, pos)
		switch spec := colour.(type) {
		case color.YCbCr:
			y = int(spec.Y)
			cb = int(spec.Cb)
			cr = int(spec.Cr)
		default:
			r, g, b, _ := colour.RGBA()
			y8, cb8, cr8 := color.RGBToYCbCr(uint8(r), uint8(g), uint8(b))
			y = int(y8)
			cb = int(cb8)
			cr = int(cr8)
		}
		pixels[0][pos] = y
		averages[0] += y
		if pos&1 == 0 {
			prevCb = cb
			prevCr = cr
		} else {
			cb = (prevCb + cb) >> 1
			cr = (prevCr + cr) >> 1
			pixels[1][pos>>1] = cb
			pixels[2][pos>>1] = cr
			averages[1] += cb
			averages[2] += cr
		}
	}

	// Create bit vector.
	for bit := uint(0); bit < 32; bit++ { // Y first.
		if pixels[0][bit] >= averages[0]>>5 { // ">> 5" is like "/ 32".
			bits |= 1 << bit
		}
	}
	for bit := uint(0); bit < 16; bit++ { // Cb, Cr next.
		if pixels[1][bit] >= averages[1]>>4 { // ">> 4" is like "/ 16".
			bits |= 1 << (bit + 32)
		}
		if pixels[2][bit] >= averages[2]>>4 {
			bits |= 1 << (bit + 48)
		}
	}

	return
}
