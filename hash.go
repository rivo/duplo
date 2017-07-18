package duplo

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/nfnt/resize"
	"github.com/rivo/duplo/haar"
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

	// DHash is a 128 bit vector where each bit value depends on the monotonicity
	// of two adjacent pixels. The first 64 bits are based on a 8x8 version of
	// the Y colour channel. The other two 32 bits are each based on a 8x4 version
	// of the Cb, and Cr colour channel, respectively.
	DHash [2]uint64

	// Histogram is histogram quantized into 64 bits (32 for Y and 16 each for
	// Cb and Cr). A bit is set to 1 if the intensity's occurence count is large
	// than the median (for that colour channel) and set to 0 otherwise.
	Histogram uint64

	// HistoMax is the maximum value of the histogram (for each channel Y, Cb,
	// and Cr).
	HistoMax [3]float32
}

// CreateHash calculates and returns the visual hash of the provided image as
// well as a resized version of it (ImageScale x ImageScale) which may be
// ignored if not needed anymore.
func CreateHash(img image.Image) (Hash, image.Image) {
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

	// Create the dHash bit vector.
	d := dHash(img)

	// Create histogram bit vector.
	h, hm := histogram(img)

	return Hash{haar.Matrix{
		Coefs:  matrix.Coefs,
		Width:  ImageScale,
		Height: ImageScale,
	}, thresholds, ratio, d, h, hm}, scaled
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
	var thresholds haar.Coef
	for index := range thresholds {
		thresholds[index] = coefThreshold(coefs, k, index)
	}

	return thresholds
}

// ycbcr returns the YCbCr values for the given colour, converting to them if
// necessary.
func ycbcr(colour color.Color) (y, cb, cr uint8) {
	switch spec := colour.(type) {
	case color.YCbCr:
		return spec.Y, spec.Cb, spec.Cr
	default:
		r, g, b, _ := colour.RGBA()
		return color.RGBToYCbCr(uint8(r), uint8(g), uint8(b))
	}
}

// dHash computes a 128 bit vector by comparing adjacent pixels of a downsized
// version of img. The first 64 bits correspond to a 8x8 version of the Y colour
// channel. A bit is set to 1 if a pixel value is higher than that of its left
// neighbour (the first bit is 1 if its colour value is > 0.5). The other two 32
// bits correspond to the Cb and Cr colour channels, based on a 8x4 version
// each.
func dHash(img image.Image) (bits [2]uint64) {
	// Resize the image to 9x8.
	scaled := resize.Resize(8, 8, img, resize.Bicubic)

	// Scan it.
	yPos := uint(0)
	cbPos := uint(0)
	crPos := uint(32)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			yTR, cbTR, crTR := ycbcr(scaled.At(x, y))
			if x == 0 {
				// The first bit is a rough approximation of the colour value.
				if yTR&0x80 > 0 {
					bits[0] |= 1 << yPos
					yPos++
				}
				if y&1 == 0 {
					_, cbBR, crBR := ycbcr(scaled.At(x, y+1))
					if (cbBR+cbTR)>>1&0x80 > 0 {
						bits[1] |= 1 << cbPos
						cbPos++
					}
					if (crBR+crTR)>>1&0x80 > 0 {
						bits[1] |= 1 << crPos
						crPos++
					}
				}
			} else {
				// Use a rough first derivative for the other bits.
				yTL, cbTL, crTL := ycbcr(scaled.At(x-1, y))
				if yTR > yTL {
					bits[0] |= 1 << yPos
					yPos++
				}
				if y&1 == 0 {
					_, cbBR, crBR := ycbcr(scaled.At(x, y+1))
					_, cbBL, crBL := ycbcr(scaled.At(x-1, y+1))
					if (cbBR+cbTR)>>1 > (cbBL+cbTL)>>1 {
						bits[1] |= 1 << cbPos
						cbPos++
					}
					if (crBR+crTR)>>1 > (crBL+crTL)>>1 {
						bits[1] |= 1 << crPos
						crPos++
					}
				}
			}
		}
	}

	return
}

// histogram calculates a histogram based on the YCbCr values of img and returns
// a rough approximation of it in 64 bits. For each colour channel, a bit is
// set if a histogram value is greater than the median. The Y channel gets 32
// bits, the Cb and Cr values each get 16 bits.
func histogram(img image.Image) (bits uint64, histoMax [3]float32) {
	h := new([64]int)

	// Create histogram.
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			y, cb, cr := ycbcr(img.At(x, y))
			h[y>>3]++
			h[32+cb>>4]++
			h[48+cr>>4]++
		}
	}

	// Calculate medians and maximums.
	median := func(v []int) (int, float32) {
		sorted := make([]int, len(v))
		copy(sorted, v)
		sort.Ints(sorted)
		return sorted[len(v)/2], float32(sorted[len(v)-1]) /
			float32((bounds.Max.X-bounds.Min.X)*(bounds.Max.Y-bounds.Min.Y))
	}
	my, yMax := median(h[:32])
	mcb, cbMax := median(h[32:48])
	mcr, crMax := median(h[48:])
	histoMax[0] = yMax
	histoMax[1] = cbMax
	histoMax[2] = crMax

	// Quantize histogram.
	for index, value := range h {
		if index < 32 {
			if value > my {
				bits |= 1 << uint(index)
			}
		} else if index < 48 {
			if value > mcb {
				bits |= 1 << uint(index-32)
			}
		} else {
			if value > mcr {
				bits |= 1 << uint(index-32)
			}
		}
	}

	return
}
