package file

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
)

// CopyIfChanged copies source data to a destination filename if it has changed.
func CopyIfChanged(destFilename string, source []byte, perm os.FileMode) error {
	matches, err := probablyMatches(destFilename, source)
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

// probablyMatches checks, heuristically, if a file matches source data.
// To save time, it only compares the file size and the end of its contents (up to 32KB).
func probablyMatches(filename string, data []byte) (bool, error) {
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

	return matchEndOfFile(f, data, 32*1024)
}

func matchEndOfFile(f *os.File, b []byte, size int) (bool, error) {
	buf := make([]byte, min(size, len(b)))
	offset := max(0, len(b)-size)
	n, err := f.ReadAt(buf, int64(offset))
	if err != nil && err != io.EOF {
		return false, err
	}

	return bytes.Equal(b[offset:], buf[:n]), nil
}
