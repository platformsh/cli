package file

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
)

// WriteIfNeeded writes data to a destination file, only if the file does not exist or if it was partially written.
// To save time, it only checks that the file size is correct, and then matches the end of its contents (up to 32KB).
func WriteIfNeeded(destFilename string, source []byte, perm os.FileMode) error {
	matches, err := probablyMatches(destFilename, source, 32*1024)
	if err != nil || matches {
		return err
	}
	return Write(destFilename, source, perm)
}

// Write creates or overwrites a file, somewhat atomically, using a temporary file next to it.
func Write(path string, content []byte, fileMode fs.FileMode) error {
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, content, fileMode); err != nil {
		return err
	}

	return os.Rename(tmpFile, path)
}

// probablyMatches checks if a file exists and matches the end of source data (up to checkSize bytes).
func probablyMatches(filename string, data []byte, checkSize int) (bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false, err
	}
	if fi.Size() != int64(len(data)) {
		return false, nil
	}

	buf := make([]byte, min(checkSize, len(data)))
	offset := max(0, len(data)-checkSize)
	n, err := f.ReadAt(buf, int64(offset))
	if err != nil && err != io.EOF {
		return false, err
	}

	return bytes.Equal(data[offset:], buf[:n]), nil
}
