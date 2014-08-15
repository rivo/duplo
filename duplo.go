package duplo

import (
	"bytes"
	"encoding/gob"
	"image"
)

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

// Add adds an image to the store. The provided ID is the value returned by
// an similarity query. When this function returns, the image itself is not
// used or needed anymore.
func (store *Store) Add(id interface{}, image image.Image) (err error) {
	store.Size++
	store.Modified = true
	return
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
