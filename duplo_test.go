package duplo

import (
	"github.com/rivo/duplo/haar"
	"testing"
)

// Test the QuickSelect algorithm.
func TestQuickSelect(t *testing.T) {
	coefs := []haar.Coef{
		haar.Coef{1, -5},
		haar.Coef{2, 2},
		haar.Coef{3, -7.5},
		haar.Coef{4, 1},
		haar.Coef{5, 0},
		haar.Coef{6, 6},
		haar.Coef{7, -3},
		haar.Coef{8, -9},
		haar.Coef{9, 4.7},
		haar.Coef{10, 4.7},
		haar.Coef{11, 8},
		haar.Coef{12, -2.2}}

	thresholds := coefThresholds(coefs, 4)

	if thresholds[0] != 9 || thresholds[1] != 6 {
		t.Errorf("Wrong thresholds, should be [9 6], is %v", thresholds)
	}
}
