package genesyscloud

import (
	"fmt"
	"io"
)

func ResetReader(reader io.Reader) (io.Reader, error) {
	// Check if the reader supports seeking
	if seeker, ok := reader.(io.Seeker); ok {
		// Attempt to reset the reader to the beginning
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}
		// Return the original reader (which is also an io.Reader)
		return reader, nil
	}

	// Return an error if the reader does not support seeking
	return nil, fmt.Errorf("reader does not support seeking")
}
