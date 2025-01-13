package file

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

// CopyIfChanged copies source data to a destination filename if it has changed.
func CopyIfChanged(destFilename string, source []byte, perm os.FileMode) error {
	beforeCheck := time.Now()
	matches, err := probablyMatches(destFilename, source)
	if err != nil || matches {
		if matches {
			log.Printf("took %s to check file: %s", time.Since(beforeCheck), filepath.Base(destFilename))
		}
		return err
	}
	defer func() {
		log.Printf("took %s to copy file: %s", time.Since(beforeCheck), destFilename)
	}()
	return writeFile(destFilename, source, perm)
}

// writeFile creates or overwrites a file, somewhat atomically, using a temporary file next to it.
func writeFile(path string, content []byte, fileMode fs.FileMode) error {
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

	// Read the end of the file (up to 32 KB).
	buf := make([]byte, min(32*1024, len(data)))
	offset := max(0, len(data)-32*1024)
	n, err := f.ReadAt(buf, int64(offset))
	if err != nil && err != io.EOF {
		return false, err
	}

	return bytes.Equal(data[offset:], buf[:n]), nil
}
