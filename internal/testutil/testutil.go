package testutil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

const (
	// where to find test assets, relative to the root of the project
	testAssetRoot = "./testdata"
)

// ReadBytesFromTestAsset locates the requested test asset and returns its contents
func ReadBytesFromTestAsset(name string) ([]byte, error) {
	// Attempt to find asset
	path, err := FindTestAssetByName(name)
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

// ReadStringFromTestAsset locates the requested test asset and returns its contents as a string
func ReadStringFromTestAsset(name string) (string, error) {
	// Read as bytes and just cast to string
	b, err := ReadBytesFromTestAsset(name)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// FindTestAssetByName crawls the test asset directory and attempts to find an asset matching the provided name
//
// TODO(nsiow) elaborate on expectations here
func FindTestAssetByName(name string) (string, error) {
	testAssetDir, err := findTestAssetDirectory()
	if err != nil {
		return "", err
	}

	// Walk the directory and save matches
	var matches []string
	err = filepath.WalkDir(testAssetDir, func(path string, d fs.DirEntry, err error) error {
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

// findTestAssetDirectory walks up the package tree from the test file until it finds our test asset directory
func findTestAssetDirectory() (string, error) {

	dir := ""
	depth := 10
	for depth > 0 {
		// Go up another level
		dir = path.Join(dir, "..")
		root := path.Join(dir, testAssetRoot)
		exists, err := dirExists(root)
		if err != nil {
			return "", fmt.Errorf("error looking for test directory at %s: %v", dir, err)
		}
		if exists {
			return dir, nil
		}

		depth--
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to determine current directory: %v", err)
	}
	return "", fmt.Errorf("unable to find test asset directory from start '%s'", cwd)
}

// TODO(nsiow) have this perform a file vs directory test as well
// dirExists determines whether or not the provided dir exists within the current directory
func dirExists(dir string) (bool, error) {
	_, err := os.Stat(dir)

	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, fs.ErrNotExist):
		return false, nil
	default:
		return false, err
	}
}
