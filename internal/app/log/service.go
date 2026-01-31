package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"workforge/internal/app/hook"
)

const (
	colReset   = "\x1b[0m"
	colDim     = "\x1b[2m"
	colRed     = "\x1b[31m"
	colGreen   = "\x1b[32m"
	colYellow  = "\x1b[33m"
	colBlue    = "\x1b[34m"
	colMagenta = "\x1b[35m"
)

const (
	iconInfo    = ""
	iconWarn    = ""
	iconError   = ""
	iconSuccess = ""
	iconDebug   = ""
)

type HookRunner interface {
	RunOnError(payload hook.ErrorPayload) []hook.HookResult
	RunOnWarning(payload hook.WarningPayload) []hook.HookResult
	RunOnDebug(payload hook.DebugPayload) []hook.HookResult
	RunOnMessage(payload hook.MessagePayload) []hook.HookResult
}

type LogService struct {
	hooks   HookRunner
	project string
}

func NewLogService(hooks HookRunner) *LogService {
	return &LogService{
		hooks:   hooks,
		project: projectNameFromCwd(),
	}
}

func (s *LogService) SetProject(name string) {
	s.project = strings.TrimSpace(name)
}

func (s *LogService) Error(context string, err error) {
	if err == nil {
		return
	}
	s.out(os.Stderr, colRed, iconError, "ERROR", "%v", err)

	if s.hooks != nil {
		s.hooks.RunOnError(hook.ErrorPayload{
			Error:   err.Error(),
			Context: context,
			Project: s.project,
		})
	}
}

func (s *LogService) ErrorMsg(context string, msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	s.out(os.Stderr, colRed, iconError, "ERROR", "%s", formatted)

	if s.hooks != nil {
		s.hooks.RunOnError(hook.ErrorPayload{
			Error:   formatted,
			Context: context,
			Project: s.project,
		})
	}
}

func (s *LogService) Warn(context string, msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	if Verbose() {
		s.out(os.Stderr, colYellow, iconWarn, "WARN", "%s", formatted)
	}

	if s.hooks != nil {
		s.hooks.RunOnWarning(hook.WarningPayload{
			Warning: formatted,
			Context: context,
			Project: s.project,
		})
	}
}

func (s *LogService) Info(context string, msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	s.out(os.Stdout, colBlue, iconInfo, "INFO", "%s", formatted)

	if s.hooks != nil {
		s.hooks.RunOnMessage(hook.MessagePayload{
			Message: formatted,
			Source:  context,
			Project: s.project,
		})
	}
}

func (s *LogService) Success(context string, msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	s.out(os.Stdout, colGreen, iconSuccess, "OK", "%s", formatted)

	if s.hooks != nil {
		s.hooks.RunOnMessage(hook.MessagePayload{
			Message: formatted,
			Source:  context,
			Project: s.project,
		})
	}
}

func (s *LogService) Debug(context string, msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	if Verbose() {
		s.out(os.Stdout, colMagenta, iconDebug, "DEBUG", "%s", formatted)
	}

	if s.hooks != nil {
		s.hooks.RunOnDebug(hook.DebugPayload{
			Message: formatted,
			Context: context,
			Project: s.project,
		})
	}
}

func (s *LogService) out(w *os.File, color, icon, label, msg string, args ...any) {
	if label != "" {
		fmt.Fprintf(w, "%s%s [%s]%s %s\n", color, icon, label, colReset, fmt.Sprintf(msg, args...))
		return
	}
	fmt.Fprintf(w, "%s%s%s %s\n", color, icon, colReset, fmt.Sprintf(msg, args...))
}

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
