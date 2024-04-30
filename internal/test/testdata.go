//go:build testonly

package test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	TEST_ASSET_ROOT = "./testdata"
)

// findTestAssetByName crawls the test asset directory and attempts to find an asset matching the provided name
//
// TODO(nsiow) elaborate on expectations here
func findTestAssetByName(name string) (string, error) {
	// Walk the directory and save matches
	var matches []string
	err := filepath.WalkDir(TEST_ASSET_ROOT, func(path string, d fs.DirEntry, err error) error {
		// Propagate error if we encountered one
		if err != nil {
			return err
		}

		// Check for the basename of the test asset; sans path and root
		fn := filepath.Base(path)
		ext := filepath.Ext(fn)
		base := fn[0 : len(fn)-len(ext)]

		// If it matches, add it
		if base == name {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// We expect exactly 1 match; provide an appropriate error otherwise
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("unable to find file matching name: '%s'", name)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("too many matches for '%s': %v", name, matches)
	}
}

// readBytesFromTestAsset locates the requested test asset and returns its contents
func readBytesFromTestAsset(name string) ([]byte, error) {
	// Attempt to find asset
	path, err := findTestAssetByName(name)
	if err != nil {
		return nil, fmt.Errorf("unable to find asset: %v", err)
	}

	// Read and return data
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read asset '%s': %v", path, err)
	}
	return data, nil
}

// readStringFromTestAsset locates the requested test asset and returns its contents as a string
func readStringFromTestAsset(name string) (string, error) {
	// Read as bytes and just cast to string
	b, err := readBytesFromTestAsset(name)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
