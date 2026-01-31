package cli

import (
	"os"
	"path/filepath"

	applog "workforge/internal/app/log"
)

type LogService = applog.LogService

func projectNameFromCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	absPath, err := filepath.Abs(cwd)
	if err != nil {
		return filepath.Base(cwd)
	}
	return filepath.Base(absPath)
}
