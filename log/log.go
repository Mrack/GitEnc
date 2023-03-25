package gitenc

import (
	"log"
)

const (
	ANSI_RED     = "\033[31m"
	ANSI_YELLOW  = "\033[33m"
	ANSI_BLUE    = "\033[34m"
	ANSI_MAGENTA = "\033[35m"
	ANSI_CYAN    = "\033[36m"
	ANSI_RESET   = "\033[0m"
)

type LogLevel int

const (
	logLevelInfo LogLevel = iota
	logLevelWarning
	logLevelError
)

func InitLog() {
	log.SetFlags(0)
	// Set the log output to os.Stderr to separate errors from normal logs
	log.SetOutput(log.Writer())
}

func getLogLevelPrefix(logLevel LogLevel) string {
	switch logLevel {
	case logLevelWarning:
		return ANSI_YELLOW + "[WARNING]" + ANSI_RESET
	case logLevelError:
		return ANSI_RED + "[ERROR]" + ANSI_RESET
	default:
		return ANSI_BLUE + "[INFO]" + ANSI_RESET
	}
}

func getLogLevel(s string) LogLevel {
	if s == logLevelWarning.String() {
		return logLevelWarning
	} else if s == logLevelError.String() {
		return logLevelError
	}
	return logLevelInfo
}

func (ll LogLevel) String() string {
	switch ll {
	case logLevelInfo:
		return "INFO"
	case logLevelWarning:
		return "WARNING"
	case logLevelError:
		return "ERROR"
	default:
		return ""
	}
}

func Info(msg ...any) {
	log.Println(append([]any{getLogLevelPrefix(logLevelInfo)}, msg...)...)
}

func Log(msg ...any) {
	log.Println(append([]any{"\t"}, msg...)...)
}

func Warning(msg ...any) {
	log.Println(append([]any{getLogLevelPrefix(logLevelWarning)}, msg...)...)
}

func Error(msg ...any) {
	log.Println(append([]any{getLogLevelPrefix(logLevelError)}, msg...)...)
}
