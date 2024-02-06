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
		return reader, nil
	}

	return nil, fmt.Errorf("reader does not support seeking")
}
