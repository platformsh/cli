package legacy

type phpManager interface {
	// copy writes embedded PHP files to temporary files.
	copy() error

	// binPath returns the path to the temporary PHP binary.
	binPath() string

	// iniSettings returns PHP INI entries (key=value format).
	iniSettings() []string
}

type phpManagerPerOS struct {
	cacheDir string
}

func newPHPManager(cacheDir string) phpManager {
	return &phpManagerPerOS{cacheDir}
}
