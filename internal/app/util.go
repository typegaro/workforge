package app

import (
	"os"
	"strings"
)

func RepoURLToName(url string) string {
	parts := strings.Split(url, "/")
	return strings.TrimSuffix(parts[len(parts)-1], ".git")
}

func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if info, err := os.Stat(src); err == nil {
		mode = info.Mode()
	}
	return os.WriteFile(dst, data, mode)
}
