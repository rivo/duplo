package duplo

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/rivo/duplo/haar"
)

const (
	// First JPEG.
	imgA = "/9j/4AAQSkZJRgABAQAAAQABAAD//gA7Q1JFQVRPUjogZ2QtanBlZyB2MS4wICh1c2luZyBJSkc" +
		"gSlBFRyB2NjIpLCBxdWFsaXR5ID0gNzUK/9sAQwAIBgYHBgUIBwcHCQkICgwUDQwLCwwZEhMPFB" +
		"0aHx4dGhwcICQuJyAiLCMcHCg3KSwwMTQ0NB8nOT04MjwuMzQy/9sAQwEJCQkMCwwYDQ0YMiEcI" +
		"TIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIy/8AAEQgA" +
		"MgAyAwEiAAIRAQMRAf/EAB8AAAEFAQEBAQEBAAAAAAAAAAABAgMEBQYHCAkKC//EALUQAAIBAwM" +
		"CBAMFBQQEAAABfQECAwAEEQUSITFBBhNRYQcicRQygZGhCCNCscEVUtHwJDNicoIJChYXGBkaJS" +
		"YnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6g4SFhoeIiYqSk5SVl" +
		"peYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2drh4uPk5ebn6Onq8fLz9PX2" +
		"9/j5+v/EAB8BAAMBAQEBAQEBAQEAAAAAAAABAgMEBQYHCAkKC//EALURAAIBAgQEAwQHBQQEAAE" +
		"CdwABAgMRBAUhMQYSQVEHYXETIjKBCBRCkaGxwQkjM1LwFWJy0QoWJDThJfEXGBkaJicoKSo1Nj" +
		"c4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoKDhIWGh4iJipKTlJWWl5iZmqKjp" +
		"KWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uLj5OXm5+jp6vLz9PX29/j5+v/aAAwD" +
		"AQACEQMRAD8A8JVPMK5+8f1qZ7ZlkAweQMCk2FWA64IrpXltbhB5UQWZQoA/vHaM/rUX1H0M2w0" +
		"yG5dkZ8Oo4XoSfb6VBLCFk8uZsITjcO9X54wgUYOwfeJznP4Ve02JLySVREgVFyxkOCB9TTuQ9D" +
		"JvtMgjUPA5KAfMTxn6UxNKka281E+XHBbjNdFN4faGzhvpp1MTk+XGeSME9K14jp0WgSTSskZwV" +
		"Xew3E46+2f0qebQo83MUmTmVxRVl/s5kY+Yep7GiquBL9n3TSDB+X0+tdK2nWcaiSMz7jkkPFwP" +
		"pgntXLQ3nkxu5PmFjxnrXb+FbbTr54PJlktSMM6hGkJPXIABrNvl1ZW+iMiOEXEpPmqFA+VicZ/" +
		"OrohFpFcJDbvOdqtnaQv1J/GuzWHT7CZ4lbMSn5GlUqenPVfqa2ftVldwXIjWDLxKuPM5GMc8gV" +
		"Dq3GoPseeala69f+HbK4lSOC1clY0jU5yOD/kVc0nwNaT6De3d61w9zGhKdSB8rHoOeoHeuz1DW" +
		"/s+jQWn9nW7ohwsnmKcn86x4/F4tLO5tDaIsjr14YE898+9KMvuG0eQtZYY/L3oro2nbef9GXr6" +
		"j/GitOYn5HFxuchfSus0q1nWATNPtT+FM8n/AArlYQF+Y9+g9a0ftsqwlVJPrg05pvYqDS3O3to" +
		"FuAd10A47A1q22ledG5+07QBjJNeb21/J5gKOVcdDV+TX7oDaHII/I1k4yLUkd7L/AGfY2+JQJL" +
		"heglXMcnv7msKTUYplcpZ21tKf4o0Ck/QgVhLrcl1GVmYsgH3c0rX8M0PyNuYfdDc4/wDr0lF9R" +
		"troQPdTeY372Tqe9FUTMcnmitTMyj/rG+tWIyaKKtkoY3E5xxxT3Y7c5NFFIZGCeDk5zUzMflOT" +
		"miigEXkVSikgEkcnFFFFIZ//2Q=="

	// Second JPEG, different from imgA.
	imgB = "/9j/4AAQSkZJRgABAQAAAQABAAD//gA7Q1JFQVRPUjogZ2QtanBlZyB2MS4wICh1c2luZyBJSkc" +
		"gSlBFRyB2NjIpLCBxdWFsaXR5ID0gNzUK/9sAQwAIBgYHBgUIBwcHCQkICgwUDQwLCwwZEhMPFB" +
		"0aHx4dGhwcICQuJyAiLCMcHCg3KSwwMTQ0NB8nOT04MjwuMzQy/9sAQwEJCQkMCwwYDQ0YMiEcI" +
		"TIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIy/8AAEQgA" +
		"MgAyAwEiAAIRAQMRAf/EAB8AAAEFAQEBAQEBAAAAAAAAAAABAgMEBQYHCAkKC//EALUQAAIBAwM" +
		"CBAMFBQQEAAABfQECAwAEEQUSITFBBhNRYQcicRQygZGhCCNCscEVUtHwJDNicoIJChYXGBkaJS" +
		"YnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6g4SFhoeIiYqSk5SVl" +
		"peYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2drh4uPk5ebn6Onq8fLz9PX2" +
		"9/j5+v/EAB8BAAMBAQEBAQEBAQEAAAAAAAABAgMEBQYHCAkKC//EALURAAIBAgQEAwQHBQQEAAE" +
		"CdwABAgMRBAUhMQYSQVEHYXETIjKBCBRCkaGxwQkjM1LwFWJy0QoWJDThJfEXGBkaJicoKSo1Nj" +
		"c4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoKDhIWGh4iJipKTlJWWl5iZmqKjp" +
		"KWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uLj5OXm5+jp6vLz9PX29/j5+v/aAAwD" +
		"AQACEQMRAD8A8O0+1a9u44VTdI7YAzjdXax/DnVZ9DGrJp8otWQurAg4Hvznt6VkeB4UPiSxeTO" +
		"0TKePrX0zYXcMHgxrUNnbCyqOMdCK4MRXcZ8q7DSPlu90GSzUiVhE4/5ZyAq1Y00DR9cY9RXtnj" +
		"GUa1seXYsqW833VwOi/wCFea6zZiO1PyYIkYZxVUMQ5pXG4nLEU5Bggn8B61IY9vLdOw9ab1bJr" +
		"sJJg7Y/1h/OioAOP/r0UgOt8GkQ6tbyH+GRT+texnUrg6RPGivsXzgSBwOvH614poVwLZ/M3YYc" +
		"g+9dda6nrssD+Q7lHy7cnBOeT6VwVqalPmZadkSalLcBommDqGt32s+QGBXPFc7rdws9mU2H5W3" +
		"dOua1LS21PVVeM38aGI7drrlhj2x0rN1N9SsmeKeZdwPVVBDfpx9DVQjBSsnqguzjnySc9aaByK" +
		"0rue5u49srb0BB4UdQMVnbdrV2JkkGD60U/A9T+VFO4ja08xhgX9a7+y1TyrNIokyrKQqlj834Z" +
		"yBXnNuFUBjn2HrWkNRli8vDABR24rlq0+cpMlvbm4ivzdJI6y9ypx/L2prajJcIcsMd8niqF1P5" +
		"j7ixOfzqOJA2BI22PrzVcitdgPdFky0ZKyDrjoaoygsSSQfccVfmVGXEX3V74xVW5O+PA+9/OtI" +
		"sRUwaKQbgMZ/WirAvHqfrUh6j6UUVLARQDESeuetTKAQuR2ooqWAlzwVxxxVOfoPpRRTiA9FUop" +
		"Kgkjk4ooorQR//2Q=="

	// Third JPEG, different but visually similar to imgB.
	imgC = "/9j/4AAQSkZJRgABAQAAAQABAAD//gA7Q1JFQVRPUjogZ2QtanBlZyB2MS4wICh1c2luZyBJSkc" +
		"gSlBFRyB2NjIpLCBxdWFsaXR5ID0gNzUK/9sAQwAIBgYHBgUIBwcHCQkICgwUDQwLCwwZEhMPFB" +
		"0aHx4dGhwcICQuJyAiLCMcHCg3KSwwMTQ0NB8nOT04MjwuMzQy/9sAQwEJCQkMCwwYDQ0YMiEcI" +
		"TIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIy/8AAEQgA" +
		"MgAyAwEiAAIRAQMRAf/EAB8AAAEFAQEBAQEBAAAAAAAAAAABAgMEBQYHCAkKC//EALUQAAIBAwM" +
		"CBAMFBQQEAAABfQECAwAEEQUSITFBBhNRYQcicRQygZGhCCNCscEVUtHwJDNicoIJChYXGBkaJS" +
		"YnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6g4SFhoeIiYqSk5SVl" +
		"peYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2drh4uPk5ebn6Onq8fLz9PX2" +
		"9/j5+v/EAB8BAAMBAQEBAQEBAQEAAAAAAAABAgMEBQYHCAkKC//EALURAAIBAgQEAwQHBQQEAAE" +
		"CdwABAgMRBAUhMQYSQVEHYXETIjKBCBRCkaGxwQkjM1LwFWJy0QoWJDThJfEXGBkaJicoKSo1Nj" +
		"c4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoKDhIWGh4iJipKTlJWWl5iZmqKjp" +
		"KWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uLj5OXm5+jp6vLz9PX29/j5+v/aAAwD" +
		"AQACEQMRAD8A8HWPJX3qw9syyAYPIGBSbSrAYzgiukeW1mQeVEFlUKAPU7Rn9ai+o+hn2OmQ3Ls" +
		"jPh1HC9CT7fSq8sIWTy5mAQtjcO9aE8YQKMHYPvE5zn8Ku6bCl5JKoiQKi5ZpDggfU07kPQyb7T" +
		"II1DwSExgfMSMZ+lRrpcrW5kjTC4+83H5V0c/h9obOG9lnUxPny0PJGCela8f9nQ+HpJpGSPgqo" +
		"dhuJx156ZqebQo81NucnLtmirTi3MjHe3U9jRVATfZ900g2n5fT610klhYwoJFacFjyHi4/DBNc" +
		"vDeeTG7k+YWPGetdl4b03StVmtWSSW1aM7mVVZ8n1wAazb5dWVvojNjtzNMR5gCjoTkfzq6Ivss" +
		"VwlvbvOSFbIUhfqfXrXXfYtKs7qREUeWD8rSblPTB6j61sl9Pure5EaQfPEiY805GMc84qHVuNQ" +
		"fY4LU7PXr7w9YzzLHBauWWOONTnIODmrel+BrSbw9e3l41w9zGhKckgfKx7e4HeuxvtZFposNmu" +
		"nWzRocLIJFOT/31WTH4vFrZXNp9kRZHXg8MCee+felGX3DaPIWssORt70V0bTvvb/Rl6+o/xorX" +
		"mJ+RxcbnIX0rq9LtZxB5pn2p2UHn/wCtXKRcGtMXkiwbUJI74NE03sVBpbnb2sCTqd10A47A1rW" +
		"2lCaNz9p2gDGSa83tb+TzAVkKsOhq9Jr10BtDkH9DWThItSR3sp0/T7bEih7hegmXMb+/vWFJqM" +
		"cyuUs7a2lP8UaBSfoQKwl1uS5jKzMWQD7maR7+GSH5Dux0Dc7aSi+o210IXuZfMb97J1PeiqZmO" +
		"TzRWpmZCVajP86KKtkojbic4p7k7c5NFFIZGCeuTnNTMxwpyc+tFFAC5PrRRRQM/9k="
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
		haar.Coef{12, -2.2},
	}

	thresholds := coefThresholds(coefs, 4)

	if thresholds[0] != 9 || thresholds[1] != 6 {
		t.Errorf("Wrong thresholds, should be [9 6], is %v", thresholds)
	}
}

// Test adding an almost black image to a store.
func TestAddBasic(t *testing.T) {
	store := New()
	//image := image.NewYCbCr(image.Rect(0, 0, 100, 100), image.YCbCrSubsampleRatio444)
	//image := image.NewGray(image.Rect(0, 0, 100, 100))
	frame := image.Rect(0, 0, 100, 100)
	plate := image.NewUniform(color.RGBA{3, 0, 4, 255})
	img := image.NewRGBA(frame)
	draw.Draw(img, frame, plate, image.Point{0, 0}, draw.Over)
	hash, _ := CreateHash(img)
	id := struct{ group, file string }{"A", "12345"}
	store.Add(id, hash)

	// We have a store of one (uniform) image. Perform tests to confirm the store
	// has been built properly.
	if size := len(store.candidates); size != 1 {
		t.Errorf("Store has %d candidates, 1 expected", size)
		return
	}
	cand := store.candidates[0]
	if cand.id != id {
		t.Errorf("Wrong candidate ID, expected %v, is %v", id, cand.id)
	}
	expected := haar.Coef{0.67785, 0.251048, 0.939454}
	t.Logf("Candidate: %v", cand)
	if size := len(cand.scaleCoef); size != 3 {
		t.Errorf("Wrong scaling function coefficient size, expected 3, is %d", size)
		return
	}
	for index := range cand.scaleCoef {
		if math.Abs(expected[index]-cand.scaleCoef[index]) >= 0.000001 {
			t.Errorf("Scaling function coefficient mismatch, expected %v, is %v", expected, cand.scaleCoef)
			break
		}
	}
	// We occupy all 2499 indices because with all zeroes, there is no
	// "top 40" and thus all zeros are saved.
	for sign := 0; sign <= 1; sign++ {
		for coefIndex := 0; coefIndex < ImageScale*ImageScale; coefIndex++ {
			for colourIndex := 0; colourIndex < haar.ColourChannels; colourIndex++ {
				none := sign > 0 || coefIndex == 0
				location := sign*ImageScale*ImageScale*haar.ColourChannels + coefIndex*haar.ColourChannels + colourIndex
				indexList := store.indices[location]
				if none {
					if len(indexList) != 0 {
						t.Errorf("Non-empty index list found for sign %d, coefficient %d, colour %d: %v", sign, coefIndex, colourIndex, indexList)
						return
					}
				}
				if !none {
					if len(indexList) != 1 {
						t.Errorf("Wrong/zero size index list found for sign %d, coefficient %d, colour %d, should be length 1: %v", sign, coefIndex, colourIndex, indexList)
						return
					}
					for i, index := range indexList {
						if index != 0 {
							t.Errorf("Wrong index found for sign %d, coefficient %d, colour %d, position %d, should be 0, is %d", sign, coefIndex, colourIndex, i, index)
							return

						}
					}
				}
			}
		}
	}
}

// Test querying with real images.
func TestQuery(t *testing.T) {
	addA, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgA)))
	addB, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgB)))
	query, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgC)))

	store := New()
	hashA, _ := CreateHash(addA)
	hashB, _ := CreateHash(addB)
	store.Add("imgA", hashA)
	store.Add("imgB", hashB)

	// Some plausibility checks.
	coefCount := 0
	for _, indices := range store.indices {
		coefCount += len(indices)
	}
	if coefCount != 2*(TopCoefs-1)*3 {
		t.Errorf("Unexpected number of bucket indices, %d instead of %d", coefCount, 2*TopCoefs*3)
	}

	// Query the store.
	queryHash, _ := CreateHash(query)
	matches := store.Query(queryHash)
	sort.Sort(matches)
	if len(matches) == 0 {
		t.Errorf("Invalid query result set size, expected 0, is %d", len(matches))
		return
	}
	if matches[0].ID != "imgA" {
		t.Errorf("Query found %s but should have found imgA", matches[0].ID)
	}
}

// Test the delete function.
func TestDelete(t *testing.T) {
	store := New()

	// Add some images.
	addA, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgA)))
	addB, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgB)))
	hashA, _ := CreateHash(addA)
	hashB, _ := CreateHash(addB)
	store.Add("imgA", hashA)
	store.Add("imgB", hashB)

	// Delete one.
	store.Delete("imgA")

	// Query should only return imgB.
	query, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgC)))
	queryHash, _ := CreateHash(query)
	matches := store.Query(queryHash)
	if len(matches) != 1 {
		t.Errorf("Invalid query result set size, expected 1, is %d", len(matches))
		return
	}
	if matches[0].ID != "imgB" {
		t.Errorf("Query found %s but should have found imgB", matches[0].ID)
	}
}

// Test the ID enumeration function.
func TestIDs(t *testing.T) {
	store := New()

	// Add some images.
	addA, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgA)))
	addB, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgB)))
	hashA, _ := CreateHash(addA)
	hashB, _ := CreateHash(addB)
	store.Add("imgA", hashA)
	store.Add("imgB", hashB)

	// A comparison helper.
	compare := func(expected, computed []interface{}) bool {
		if len(computed) != len(expected) {
			t.Errorf("Invalid ID slice length, expected %d, is %d", len(expected), len(computed))
			return true
		}
	ExpLoop:
		for _, id := range expected {
			for _, checkID := range computed {
				if id == checkID {
					continue ExpLoop
				}
			}
			t.Errorf("ID %s not found", id)
			return true
		}
		return false
	}

	// First enumeration.
	ids := store.IDs()
	if compare([]interface{}{"imgA", "imgB"}, ids) {
		return
	}

	// Delete one.
	store.Delete("imgA")

	// Second enumeration.
	ids = store.IDs()
	compare([]interface{}{"imgB"}, ids)
}

// Test the ID exchange function.
func TestExchange(t *testing.T) {
	store := New()

	// Add some images.
	addA, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgA)))
	addB, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgB)))
	hashA, _ := CreateHash(addA)
	hashB, _ := CreateHash(addB)
	store.Add("imgA", hashA)
	store.Add("imgB", hashB)

	// Test failure to find original ID.
	if err := store.Exchange("does not exist", "is irrelevant"); err != nil {
		t.Errorf("Exchange returned with unexpected error message: %s", err)
		return
	}
	if len(store.ids) != 2 {
		t.Errorf("Failed exchange modified store length, expected 2, is %d", len(store.ids))
		return
	}

	// Test failure to rename into existing ID.
	if err := store.Exchange("imgA", "imgB"); err == nil {
		t.Error("Exchange into existing ID did not fail")
		return
	}

	// Now rename and check result.
	if err := store.Exchange("imgA", "imgC"); err != nil {
		t.Error(err)
		return
	}
	if len(store.ids) != 2 {
		t.Errorf("Successful exchange modified store length, expected 2, is %d", len(store.ids))
		return
	}
	if _, ok := store.ids["imgA"]; ok {
		t.Error(`ID "imgA" still found despite exchange`)
		return
	}
	if _, ok := store.ids["imgC"]; !ok {
		t.Error(`ID "imgC" not found despite exchange`)
		return
	}
	if store.candidates[0].id != "imgC" {
		t.Errorf(`Candidate 0 ID is not "imgC" but "%s"`, store.candidates[0].id)
		return
	}
	if _, ok := store.ids["imgB"]; !ok {
		t.Error(`ID "imgB" not found after exchange`)
	}
}

// Used in the next test.
type testID struct {
	Asset  string
	Number int
}

// Test serialization.
func TestGob(t *testing.T) {
	addA, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgA)))
	addB, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgB)))
	addC, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgC)))

	store := New()
	hashA, _ := CreateHash(addA)
	hashB, _ := CreateHash(addB)
	hashC, _ := CreateHash(addC)
	store.Add(testID{"image", 1}, hashA)
	store.Add(testID{"image", 2}, hashB)
	store.Add(testID{"image", 3}, hashC)

	// Serialize store.
	var file bytes.Buffer
	encoder := gob.NewEncoder(&file)
	if err := encoder.Encode(store); err != nil {
		t.Errorf("Encoding store failed: %s", err)
		return
	}

	// Unserialize store.
	var storeReloaded Store
	decoder := gob.NewDecoder(&file)
	if err := decoder.Decode(&storeReloaded); err != nil {
		t.Errorf("Decoding store failed: %s", err)
		return
	}

	// Are the candidates the same?
	if len(store.candidates) != len(storeReloaded.candidates) {
		t.Error("Candidate length not identical")
		return
	}
	for index, candidate := range store.candidates {
		if storeReloaded.candidates[index].id.(testID) != candidate.id.(testID) {
			t.Errorf("Candidate ID not identical: %v vs %v", storeReloaded.candidates[index].id, candidate.id)
			break
		}
		if len(storeReloaded.candidates[index].scaleCoef) != len(candidate.scaleCoef) {
			t.Errorf("Candidate scaling function coefficient size not identical: %d vs %d", len(storeReloaded.candidates[index].scaleCoef), len(candidate.scaleCoef))
			break
		}
		for i, v := range storeReloaded.candidates[index].scaleCoef {
			if v != candidate.scaleCoef[i] {
				t.Errorf("Candidate scaling function coefficient not identical: %v vs %v", storeReloaded.candidates[index].scaleCoef, candidate.scaleCoef)
				break
			}
		}
		if storeReloaded.candidates[index].ratio != candidate.ratio {
			t.Errorf("Candidate ratio not identical: %f vs %f", storeReloaded.candidates[index].ratio, candidate.ratio)
			break
		}
	}

	// Are the indices the same?
	if l1, l2 := len(store.indices), len(storeReloaded.indices); l1 != l2 {
		t.Errorf("Index number of signs not identical: %d vs %d", l1, l2)
		return
	}
	for location, indices := range storeReloaded.indices {
		if l1, l2 := len(indices), len(store.indices[location]); l1 != l2 {
			t.Errorf("Reloaded index slice at %d is of length %d, expected %d", location, l1, l2)
			return
		}
		for i, index := range indices {
			if index != store.indices[location][i] {
				t.Errorf("Reloaded index at %d[%d] is %d, expected %d", location, i, index, store.indices[location][i])
				return
			}
		}
	}
}

// Package example.
func Example() {
	// Create some example JPEG images.
	addA, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgA)))
	addB, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgB)))
	query, _ := jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgC)))

	// Create the store.
	store := New()

	// Turn two images into hashes and add them to the store.
	hashA, _ := CreateHash(addA)
	hashB, _ := CreateHash(addB)
	store.Add("imgA", hashA)
	store.Add("imgB", hashB)

	// Query the store for our third image (which is most similar to "imgA").
	queryHash, _ := CreateHash(query)
	matches := store.Query(queryHash)
	fmt.Println(matches[0].ID)
	// Output: imgA
}
