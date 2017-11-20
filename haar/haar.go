/*
Package haar provides a Haar wavelet function for bitmap images.
*/
package haar

import (
	"image"
	"image/color"
	"math"
)

// ColourChannels is the number of channels for one color. We will be using
// three colour channels per pixel at all times.
const ColourChannels = 3

// Coef is the union of coefficients for all channels of the original image.
type Coef [ColourChannels]float64

// Add adds another coefficient in place.
func (coef *Coef) Add(offset Coef) {
	for index := range coef {
		coef[index] += offset[index]
	}
}

// Subtract subtracts another coefficient in place.
func (coef *Coef) Subtract(offset Coef) {
	for index := range coef {
		coef[index] -= offset[index]
	}
}

// Divide divides all elements of the coefficient by a value, in place.
func (coef *Coef) Divide(value float64) {
	factor := 1.0 / value
	for index := range coef {
		coef[index] *= factor // Slightly faster.
	}
}

// Matrix is the result of the Haar transform, a two-dimensional matrix of
// coefficients.
type Matrix struct {
	// Coefs is the slice of coefficients resulting from a forward 2D Haar
	// transform. The position of a coefficient (x,y) is (y * Width + x).
	Coefs []Coef

	// The number of columns in the matrix.
	Width uint

	// The number of rows in the matrix.
	Height uint
}

// colorToCoef converts a native Color type into a YIQ Coef. We are using
// YIQ because we only have weights for them. (Apart from the score weights,
// the store is built to handle different sized Coef's so any length may be
// returned.)
func colorToCoef(gen color.Color) Coef {
	// Convert into YIQ. (We may want to convert from YCbCr directly one day.)
	r32, g32, b32, _ := gen.RGBA()
	r, g, b := float64(r32>>8), float64(g32>>8), float64(b32>>8)
	return Coef{
		(0.299900*r + 0.587000*g + 0.114000*b) / 0x100,
		(0.595716*r - 0.274453*g - 0.321263*b) / 0x100,
		(0.211456*r - 0.522591*g + 0.311135*b) / 0x100}
}

// Transform performs a forward 2D Haar transform on the provided image after
// converting it to YIQ space.
func Transform(img image.Image) Matrix {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	if width > 2 {
		// We can't handle odd widths.
		width = width &^ 1
	}
	if height > 2 {
		// We can't handle odd heights.
		height = height &^ 1
	}
	matrix := Matrix{
		Coefs:  make([]Coef, width*height),
		Width:  uint(width),
		Height: uint(height)}

	// Convert colours to coefficients.
	for row := bounds.Min.Y; row < bounds.Min.Y+height; row++ {
		for column := bounds.Min.X; column < bounds.Min.X+width; column++ {
			matrix.Coefs[(row-bounds.Min.Y)*width+(column-bounds.Min.X)] = colorToCoef(img.At(column, row))
		}
	}

	// Apply 1D Haar transform on rows.
	tempRow := make([]Coef, width)
	for row := 0; row < height; row++ {
		for step := width / 2; step >= 1; step /= 2 {
			for column := 0; column < step; column++ {
				high := matrix.Coefs[row*width+2*column]
				low := high
				offset := matrix.Coefs[row*width+2*column+1]
				high.Add(offset)
				low.Subtract(offset)
				high.Divide(math.Sqrt2)
				low.Divide(math.Sqrt2)
				tempRow[column] = high
				tempRow[column+step] = low
			}
			for column := 0; column < width; column++ {
				matrix.Coefs[row*width+column] = tempRow[column]
			}
		}
	}

	// Apply 1D Haar transform on columns.
	tempColumn := make([]Coef, height)
	for column := 0; column < width; column++ {
		for step := height / 2; step >= 1; step /= 2 {
			for row := 0; row < step; row++ {
				high := matrix.Coefs[(2*row)*width+column]
				low := high
				offset := matrix.Coefs[(2*row+1)*width+column]
				high.Add(offset)
				low.Subtract(offset)
				high.Divide(math.Sqrt2)
				low.Divide(math.Sqrt2)
				tempColumn[row] = high
				tempColumn[row+step] = low
			}
			for row := 0; row < height; row++ {
				matrix.Coefs[row*width+column] = tempColumn[row]
			}
		}
	}

	return matrix
}
