package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// QuickHashFields generates a SHA-256 hash of any number of objects using gob encoding
// Returns first 16 characters of hex-encoded hash for a good balance of uniqueness and length
// For large structs, consider passing pointers to reduce copying overhead.
func QuickHashFields(values ...interface{}) (string, error) {
	if len(values) == 0 {
		return "", fmt.Errorf("no values provided to hash")
	}

	// Keep track when we have any non-nil values
	hasNonNilValue := false

	// Use buffer pool for efficiency
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	enc := json.NewEncoder(buf)
	for _, val := range values {
		// Skip nil values
		if val == nil {
			continue
		}

		hasNonNilValue = true
		if err := enc.Encode(val); err != nil {
			return "", fmt.Errorf("failed to encode value: %v", err)
		}

	}

	// If all values were nil, return empty
	if !hasNonNilValue {
		return "", nil
	}

	h := sha256.New()
	h.Write(buf.Bytes())
	return hex.EncodeToString(h.Sum(nil))[:16], nil
}

// QuickHashFieldsWithDefault generates a SHA-256 hash of any non-nil objects using json encoding
// Returns first 16 characters of hex-encoded hash for a good balance of uniqueness and length
// If all values are nil, returns the provided default value instead
func QuickHashFieldsWithDefault(defaultValue string, values ...interface{}) (string, error) {
	hash, err := QuickHashFields(values...)
	if err != nil {
		return "", err
	}
	if hash == "" {
		return defaultValue, nil
	}
	return hash, nil
}
