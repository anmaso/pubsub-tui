package utils

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// JSONFile represents a JSON file in the working directory
type JSONFile struct {
	Name     string // File name without path
	Path     string // Full path to file
	Size     int64  // File size in bytes
	Modified int64  // Unix timestamp of last modification
}

// ListJSONFiles returns all JSON files in the specified directory
func ListJSONFiles(dir string) ([]JSONFile, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []JSONFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".json") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, JSONFile{
			Name:     name,
			Path:     filepath.Join(dir, name),
			Size:     info.Size(),
			Modified: info.ModTime().Unix(),
		})
	}

	// Sort by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files, nil
}

// ReadFile reads the entire contents of a file
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

