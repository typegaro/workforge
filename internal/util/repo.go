package util

import (
	"net/url"
	"os"
	"path"
	"strings"
)

func RepoUrlToName(repoURL string) string {
	parsed, err := url.Parse(repoURL)
	if err == nil {
		base := path.Base(parsed.Path)
		return strings.TrimSuffix(base, ".git")
	}
	parts := strings.Split(repoURL, "/")
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
