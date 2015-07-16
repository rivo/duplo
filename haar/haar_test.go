package haar

import (
	"image"
	"image/color"
	"math"
	"testing"
)

const epsilon = 0.002

// Whether or not the two coefficients are equal to an epsilon difference.
func equal(slice1, slice2 Coef) bool {
	for index := range slice1 {
		if math.Abs(slice1[index]-slice2[index]) > epsilon {
			return false
		}
	}
	return true
}

// Whether or not the two matrices are equal (uses equal() function).
func equalMatrices(matrix1, matrix2 Matrix) bool {
	if matrix1.Width != matrix2.Width {
		return false
	}
	if matrix1.Height != matrix2.Height {
		return false
	}
	if len(matrix1.Coefs) != len(matrix2.Coefs) {
		return false
	}
	for index := range matrix1.Coefs {
		if math.Abs(matrix1.Coefs[index][0]-matrix2.Coefs[index][0]) > epsilon {
			return false
		}
	}
	return true
}

// Converts a slice of floats to a Coefs slice as found in a one-value matrix.
func floatsToCoefs(floats []float64) []Coef {
	coefs := make([]Coef, len(floats))
	for index := range floats {
		coefs[index] = Coef{floats[index]}
	}
	return coefs
}

// Test coefficients.
func TestCoef(t *testing.T) {
	coef := Coef{1, 2, 3}
	copyCoef := coef
	if !equal(copyCoef, Coef{1, 2, 3}) {
		t.Errorf("Coef not a copy (%v instead of %v)", copyCoef, coef)
	}

	offset := Coef{2, 4, 6}
	coef.Add(offset)
	if !equal(coef, Coef{3, 6, 9}) {
		t.Errorf("Addition failed, result: %v", coef)
	}

	coef.Subtract(offset)
	if !equal(coef, Coef{1, 2, 3}) {
		t.Errorf("Subtraction failed, result: %v", coef)
	}

	coef.Divide(2)
	if !equal(coef, Coef{.5, 1, 1.5}) {
		t.Errorf("Division failed, result: %v", coef)
	}
}

// Test the proper RGB-YIQ conversion.
func TestColorConversion(t *testing.T) {
	rgb := color.RGBA{64, 0, 128, 255}
	coef := colorToCoef(rgb)
	if !equal(coef, Coef{0.131975, -0.0117025, 0.2084315}) {
		t.Errorf("Conversion failed (%v to %v)", rgb, coef)
	}
}

// Essentially a 1D Haar Wavelet test.
func TestSingleRow(t *testing.T) {
	// This is a rough approximation to a 4px by 1px YIQ image with pixels
	// .04, .02, .05, .05. Y, I, and Q all have the same value.
	input := &image.RGBA{
		Pix:    []uint8{26, 1, 16, 1, 13, 0, 8, 1, 33, 1, 20, 1, 33, 1, 20, 1},
		Stride: 16,
		Rect:   image.Rect(0, 0, 4, 1)}

	output := Transform(input)

	expected := Matrix{
		Coefs:  []Coef{Coef{0.08}, Coef{-0.02}, Coef{.02 / math.Sqrt2}, Coef{0}},
		Width:  4,
		Height: 1}

	if !equalMatrices(output, expected) {
		t.Errorf("Result not as expected. Result=%v, expected=%v", output, expected)
	}
}

// Essentially another 1D Haar Wavelet test.
func TestSingleColumn(t *testing.T) {
	// This is a rough approximation to a 1px by 4px YIQ image with pixels
	// .04, .02, .05, .05. Y, I, and Q all have the same value.
	input := &image.RGBA{
		Pix:    []uint8{26, 1, 16, 1, 13, 0, 8, 1, 33, 1, 20, 1, 33, 1, 20, 1},
		Stride: 4,
		Rect:   image.Rect(0, 0, 1, 4)}

	output := Transform(input)

	expected := Matrix{
		Coefs:  floatsToCoefs([]float64{0.08, -0.02, 0.02 / math.Sqrt2, 0}),
		Width:  1,
		Height: 4}

	if !equalMatrices(output, expected) {
		t.Errorf("Result not as expected. Result=%v, expected=%v", output, expected)
	}
}

// Basic 2D Haar Wavelet test.
func TestMatrix4x4(t *testing.T) {
	// This is a rough approximation to a 4px by 4px YIQ image with consecutive
	// pixels increasing by one each (.01, .02, .03, .04, ..., .16) and Y, I, and
	// Q having the same values.
	input := &image.RGBA{
		Pix: []uint8{
			7, 0, 4, 1, 13, 0, 8, 1, 20, 1, 12, 1, 26, 1, 16, 1,
			33, 1, 20, 1, 39, 1, 24, 1, 46, 1, 29, 1, 53, 2, 33, 1,
			59, 2, 37, 1, 66, 2, 41, 1, 72, 2, 45, 1, 79, 2, 49, 1,
			85, 3, 53, 1, 92, 3, 57, 1, 99, 3, 61, 1, 105, 3, 65, 1},
		Stride: 16,
		Rect:   image.Rect(0, 0, 4, 4)}

	output := Transform(input)

	expected := Matrix{
		Coefs: floatsToCoefs([]float64{
			.34, -.04, -math.Sqrt2 / 100, -math.Sqrt2 / 100,
			-.16, 0, 0, 0,
			-.04 * math.Sqrt2, 0, 0, 0,
			-.04 * math.Sqrt2, 0, 0, 0}),
		Width:  4,
		Height: 4}

	if !equalMatrices(output, expected) {
		t.Errorf("Result not as expected. Result=%v, expected=%v", output, expected)
	}
}
