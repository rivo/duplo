package haar

import (
	"image"
	"image/color"
	"math"
	"testing"
)

const epsilon = 0.0000001

// Whether or not the two slices are equal to an epsilon difference.
func equal(slice1, slice2 []float64) bool {
	if len(slice1) != len(slice2) {
		return false
	}
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
		if !equal(matrix1.Coefs[index], matrix2.Coefs[index]) {
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
	copyCoef := coef.Copy()
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

// Test the proper conversion of a colour.
func TestColorConversion(t *testing.T) {
	alpha := color.Alpha{255}
	alphaCoef := colorToCoef(alpha)
	if !equal(alphaCoef, Coef{255}) {
		t.Errorf("Conversion failed (%v to %v)", alpha, alphaCoef)
	}

	gray := color.Gray{64}
	grayCoef := colorToCoef(gray)
	if !equal(grayCoef, Coef{64}) {
		t.Errorf("Conversion failed (%v to %v)", gray, grayCoef)
	}

	gray16 := color.Gray16{65535}
	gray16Coef := colorToCoef(gray16)
	if !equal(gray16Coef, Coef{65535}) {
		t.Errorf("Conversion failed (%v to %v)", gray16, gray16Coef)
	}

	yCbCr := color.YCbCr{90, 60, 90}
	yCbCrCoef := colorToCoef(yCbCr)
	if !equal(yCbCrCoef, Coef{90, 60, 90}) {
		t.Errorf("Conversion failed (%v to %v)", yCbCr, yCbCrCoef)
	}

	rgba := color.RGBA{1, 128, 255, 73}
	rgbaCoef := colorToCoef(rgba)
	if !equal(rgbaCoef, Coef{1, 128, 255, 73}) {
		t.Errorf("Conversion failed (%v to %v)", rgba, rgbaCoef)
	}
}

// Essentially a 1D Haar Wavelet test.
func TestSingleRow(t *testing.T) {
	input := &image.Gray{
		Pix:    []uint8{4, 2, 5, 5},
		Stride: 4,
		Rect:   image.Rect(0, 0, 4, 1)}

	output := Transform(input)

	expected := Matrix{
		Coefs:  []Coef{Coef{8.0}, Coef{-2}, Coef{2 / math.Sqrt2}, Coef{0}},
		Width:  4,
		Height: 1}

	if !equalMatrices(output, expected) {
		t.Errorf("Result not as expected. Result=%v, expected=%v", output, expected)
	}
}

// Essentially another 1D Haar Wavelet test.
func TestSingleColumn(t *testing.T) {
	input := &image.Gray{
		Pix:    []uint8{4, 2, 5, 5},
		Stride: 1,
		Rect:   image.Rect(0, 0, 1, 4)}

	output := Transform(input)

	expected := Matrix{
		Coefs:  floatsToCoefs([]float64{8, -2, 2 / math.Sqrt2, 0}),
		Width:  1,
		Height: 4}

	if !equalMatrices(output, expected) {
		t.Errorf("Result not as expected. Result=%v, expected=%v", output, expected)
	}
}

// Basic 2D Haar Wavelet test.
func TestMatrix4x4(t *testing.T) {
	input := &image.Gray{
		Pix: []uint8{
			1, 2, 3, 4,
			5, 6, 7, 8,
			9, 10, 11, 12,
			13, 14, 15, 16},
		Stride: 4,
		Rect:   image.Rect(0, 0, 4, 4)}

	output := Transform(input)

	expected := Matrix{
		Coefs: floatsToCoefs([]float64{
			34, -4, -math.Sqrt2, -math.Sqrt2,
			-16, 0, 0, 0,
			-4 * math.Sqrt2, 0, 0, 0,
			-4 * math.Sqrt2, 0, 0, 0}),
		Width:  4,
		Height: 4}

	if !equalMatrices(output, expected) {
		t.Errorf("Result not as expected. Result=%v, expected=%v", output, expected)
	}
}
