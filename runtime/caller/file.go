package caller

import (
	"path/filepath"
)

// File identifies a full file path
type File string

// Base returns the base file name.
func (f File) Base() string {
	return filepath.Base(string(f))
}

// WithDir returns base file name with specified maximum directories.
func (f File) WithDir(max int) string {
	dir, file := filepath.Split(string(f))
	revParts := make([]string, 0, max)
	revParts = append(revParts, file)

	for i := 0; i < max; i++ {
		dir, file = filepath.Split(dir)
		if file == "" {
			break
		}
		revParts = append(revParts, file)
	}

	// reverse
	for i, j := 0, len(revParts)-1; i < j; i, j = i+1, j-1 {
		revParts[i], revParts[j] = revParts[j], revParts[i]
	}
	return filepath.Join(revParts...)
}
