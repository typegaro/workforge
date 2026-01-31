package log

type LogLevel int

const (
	LevelSilent LogLevel = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
)
