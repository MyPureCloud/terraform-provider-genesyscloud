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
func QuickHashFields(values ...interface{}) (string, error) {
	if len(values) == 0 {
		return "", fmt.Errorf("no values provided to hash")
	}

	// Use buffer pool for efficiency
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	enc := json.NewEncoder(buf)
	for _, val := range values {
		if err := enc.Encode(val); err != nil {
			return "", fmt.Errorf("failed to encode value: %v", err)
		}

	}

	h := sha256.New()
	h.Write(buf.Bytes())
	return hex.EncodeToString(h.Sum(nil))[:16], nil
}
