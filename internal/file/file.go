package file

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
)

// CopyIfChanged copies source data to a destination filename if it has changed.
// It is considered changed if its length or contents are different.
func CopyIfChanged(destFilename string, source []byte, perm os.FileMode) error {
	matches, err := compare(destFilename, source)
	if (err != nil && !os.IsNotExist(err)) || matches {
		return err
	}
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

// compare checks if a file matches the given source.
func compare(filename string, data []byte) (bool, error) {
	f, err := os.Open(filename)
	if err != nil {
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

	var (
		buf    = make([]byte, 32*1024)
		offset = 0
	)
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return false, err
		}
		if offset+n > len(data) || !bytes.Equal(data[offset:offset+n], buf[:n]) {
			return false, nil
		}
		offset += n
	}

	return offset == len(data), nil
}

// CheckHash checks if a file has the given SHA256 hash.
func CheckHash(filename, hash string) (bool, error) {
	fh, err := sha256File(filename)
	if err != nil {
		return false, err
	}
	return fh == hash, nil
}

// sha256File calculates the SHA256 hash of a file.
func sha256File(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
