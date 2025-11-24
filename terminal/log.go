package terminal

import (
    "fmt"
    "os"
    "strings"
)

// LogLevel represents verbosity threshold.
type LogLevel int

const (
    LevelSilent LogLevel = iota
    LevelError
    LevelWarn
    LevelInfo
    LevelDebug
)

var currentLevel = LevelInfo

// SetLogLevel sets the global log level.
func SetLogLevel(l LogLevel) { currentLevel = l }

// SetLogLevelFromString configures level from a string (e.g., "debug").
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

// Basic color helpers (ANSI). No external deps.
const (
    colReset   = "\x1b[0m"
    colDim     = "\x1b[2m"
    colRed     = "\x1b[31m"
    colGreen   = "\x1b[32m"
    colYellow  = "\x1b[33m"
    colBlue    = "\x1b[34m"
    colMagenta = "\x1b[35m"
)

// Nerd Font icons (requires Nerd Fonts in terminal).
const (
    iconInfo    = "" // nf-fa-info_circle
    iconWarn    = "" // nf-fa-exclamation_triangle
    iconError   = "" // nf-fa-times_circle
    iconSuccess = "" // nf-fa-check_circle
    iconDebug   = "" // nf-fa-bug
)

// internal print helpers
func out(w *os.File, color, icon, label, msg string, args ...any) {
    // Colored icon + [LABEL], then reset before message
    if label != "" {
        fmt.Fprintf(w, "%s%s [%s]%s %s\n", color, icon, label, colReset, fmt.Sprintf(msg, args...))
        return
    }
    // Fallback without label
    fmt.Fprintf(w, "%s%s%s %s\n", color, icon, colReset, fmt.Sprintf(msg, args...))
}

// Public log functions
func Debug(msg string, args ...any) {
    if currentLevel >= LevelDebug {
        out(os.Stdout, colMagenta, iconDebug, "DEBUG", msg, args...)
    }
}

func Info(msg string, args ...any) {
    if currentLevel >= LevelInfo {
        out(os.Stdout, colBlue, iconInfo, "INFO", msg, args...)
    }
}

func Success(msg string, args ...any) {
    if currentLevel >= LevelInfo { // Success follows info visibility
        out(os.Stdout, colGreen, iconSuccess, "OK", msg, args...)
    }
}

func Warn(msg string, args ...any) {
    if currentLevel >= LevelWarn {
        out(os.Stderr, colYellow, iconWarn, "WARN", msg, args...)
    }
}

func Error(msg string, args ...any) {
    if currentLevel >= LevelError {
        out(os.Stderr, colRed, iconError, "ERROR", msg, args...)
    }
}

// Step prints a dimmed step line without changing level semantics.
func Step(msg string, args ...any) {
    if currentLevel >= LevelInfo {
        out(os.Stdout, colDim, "•", "", msg, args...)
    }
}
