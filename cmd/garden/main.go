package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andresfelipemendez/garden/internal/garden"
)

const (
	notesDir  = "notes"
	publicDir = "public"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	paths, err := discoverNotes(notesDir)
	if err != nil {
		return fmt.Errorf("discover notes: %w", err)
	}

	if err := garden.Build(paths, publicDir); err != nil {
		return fmt.Errorf("build: %w", err)
	}

	return nil
}

func discoverNotes(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	return paths, nil
}
