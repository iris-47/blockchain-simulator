package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
)

// encodes the object using gob encoding.
// gob will automatically handles pointer dereferencing, so if you pass a pointer to `object`, it will encode the object that the pointer points to.
func Encode(object interface{}) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(object); err != nil {
		return nil
	}
	return buf.Bytes()
}

// decodes the data using gob encoding and stores the result in the object
func Decode(data []byte, object interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(object)
}

// hashes the object using sha256
func Hash(object interface{}) []byte {
	hash := sha256.Sum256(Encode(object))
	return hash[:]
}
