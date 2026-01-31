package log

import (
	"fmt"
	"os"
	"strings"
)

var verboseFlag = "false"

func Verbose() bool {
	return verboseFlag == "true"
}

type LogLevel int

const (
	LevelSilent LogLevel = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
)

var currentLevel = LevelInfo

func SetLogLevel(l LogLevel) { currentLevel = l }

func SetLogLevelFromString(s string) {
	s = strings.TrimSpace(strings.ToUpper(s))
	switch s {
	case "", "INFO":
		currentLevel = LevelInfo
	case "DEBUG":
		currentLevel = LevelDebug
	case "WARN", "WARNING":
		currentLevel = LevelWarn
	case "ERROR":
		currentLevel = LevelError
	case "SILENT", "QUIET":
		currentLevel = LevelSilent
	default:
		currentLevel = LevelInfo
	}
}

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

func out(w *os.File, color, icon, label, msg string, args ...any) {
	if label != "" {
		fmt.Fprintf(w, "%s%s [%s]%s %s\n", color, icon, label, colReset, fmt.Sprintf(msg, args...))
		return
	}
	fmt.Fprintf(w, "%s%s%s %s\n", color, icon, colReset, fmt.Sprintf(msg, args...))
}

func Debug(msg string, args ...any) {
	if Verbose() && currentLevel >= LevelDebug {
		out(os.Stdout, colMagenta, iconDebug, "DEBUG", msg, args...)
	}
}

func Info(msg string, args ...any) {
	if currentLevel >= LevelInfo {
		out(os.Stdout, colBlue, iconInfo, "INFO", msg, args...)
	}
}

func Success(msg string, args ...any) {
	if currentLevel >= LevelInfo {
		out(os.Stdout, colGreen, iconSuccess, "OK", msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if Verbose() && currentLevel >= LevelWarn {
		out(os.Stderr, colYellow, iconWarn, "WARN", msg, args...)
	}
}

func Error(msg string, args ...any) {
	if currentLevel >= LevelError {
		out(os.Stderr, colRed, iconError, "ERROR", msg, args...)
	}
}

func Step(msg string, args ...any) {
	if currentLevel >= LevelInfo {
		out(os.Stdout, colDim, "•", "", msg, args...)
	}
}
