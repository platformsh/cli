package legacy

type phpManager interface {
	// copy copies PHP files to the filesystem
	copy() error

	// binaryPath returns the path to the PHP binary.
	binaryPath() string

	// iniSettings returns PHP INI entries (key=value format).
	iniSettings() []string
}

type phpManagerPerOS struct {
	cacheDir string
}

func newPHPManager(cacheDir string) phpManager {
	return &phpManagerPerOS{cacheDir: cacheDir}
}
