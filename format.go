package main

import (
	"os"
	"path"

	"golang.org/x/tools/imports"
)

func handleFmt(code []byte) ([]byte, error) {
	return imports.Process(path.Join(os.TempDir(), "main.go"), code, nil)
}
