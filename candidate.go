package duplo

import (
	"github.com/rivo/duplo/haar"
)

// candidate represents an image in the store or, rather, a candidate to be
// selected as the winner in a similarity query.
type candidate struct {
	// id is the unique ID that identifies an image.
	id interface{}

	// scaleCoef is the scaling function coefficient, the coefficients at index
	// (0,0) of the Haar matrix.
	scaleCoef haar.Coef

	// ratio is image width / image height.
	ratio float64

	// The dHash bit vector (see Hash for more information).
	dHash [2]uint64

	// The histogram bit vector (see Hash for more information).
	histogram uint64

	// The histogram maximum (see Hash for more information).
	histoMax [3]float32
}
