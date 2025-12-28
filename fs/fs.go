package fs

import (
	"os"
	"path/filepath"
)

// DefaultCacheDir returns the default cache directory for diffstory.
// Uses XDG_CACHE_HOME if set, otherwise falls back to ~/.cache/diffstory.
func DefaultCacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "diffstory")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "diffstory")
}
