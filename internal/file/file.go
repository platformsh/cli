package file

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
)

const HashExt = ".sha256"

var testableFS = os.DirFS("/")

// CopyIfChanged copies source data to a destination filename if it has changed.
// It is considered changed if its length or hash are different.
// The hash may be a static hash saved alongside (with HashExt) or computed dynamically.
func CopyIfChanged(destFilename string, source []byte, sourceHash string) error {
	sizeOK, err := checkSize(destFilename, len(source))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if sizeOK {
		hashOK, err := CheckHash(destFilename, sourceHash)
		if hashOK || err != nil {
			return err
		}
	}
	if err := os.WriteFile(destFilename, source, 0o644); err != nil {
		return err
	}
	return SaveHash(destFilename, sourceHash)
}

// CheckHash checks if a file has the given SHA256 hash.
// It supports reading the file's current hash from a static file saved next to it with the HashExt extension.
func CheckHash(filename, hash string) (bool, error) {
	if fh, err := fs.ReadFile(testableFS, filename+HashExt); err == nil {
		return string(fh) == hash, nil
	}
	fh, err := sha256Sum(testableFS, filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return true, err
	}
	return fh == hash, nil
}

// SaveHash saves a hash alongside a file, with the same filename plus the HashExt extension.
func SaveHash(filename, hash string) error {
	return os.WriteFile(filename+HashExt, []byte(hash), 0o644)
}

// sha256Sum calculates the SHA256 hash of a file.
func sha256Sum(filesystem fs.FS, filename string) (string, error) {
	f, err := filesystem.Open(filename)
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

// checkSize checks if a file exists and has an exact size.
func checkSize(filename string, size int) (bool, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		return false, err
	}
	return stat.Size() == int64(size), nil
}
