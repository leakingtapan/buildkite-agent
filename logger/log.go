package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	nocolor = "0"
	red     = "31"
	green   = "1;32"
	yellow  = "33"
	blue    = "34"
	gray    = "1;30"
	cyan    = "1;36"
)

const (
	DateFormat = "2006-01-02 15:04:05"
)

var (
	mutex  = sync.Mutex{}
	colors bool
)

type Logger struct {
	Level  Level
	Colors bool
	Prefix string
	Writer io.Writer
	ExitFn func()
}

func NewLogger() *Logger {
	return &Logger{
		Level:  DEBUG,
		Colors: true,
		Writer: os.Stderr,
	}
}

func SetColors(b bool) {
	colors = b
}

func ColorsEnabled() bool {
	if runtime.GOOS == "windows" {
		// Boo, no colors on Windows.
		return false
	} else {
		// Colors can only be shown if STDOUT is a terminal
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			return colors
		} else {
			return false
		}
	}
}

func (l *LevelLogger) SetLevel(v Level) {
	l.Level = v
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.Level == DEBUG {
		l.log(DEBUG, format, v...)
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ERROR, format, v...)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(FATAL, format, v...)
	os.Exit(1)
}

func (l *Logger) Notice(format string, v ...interface{}) {
	if l.Level <= NOTICE {
		l.log(NOTICE, format, v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.Level <= INFO {
		l.log(INFO, format, v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.Level <= WARN {
		l.log(WARN, format, v...)
	}
}

func (l *Logger) log(level Level, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	now := time.Now().Format(DateFormat)
	line := ""

	if l.Colors {
		prefixColor := green
		messageColor := nocolor

		switch level {
		case DEBUG:
			prefixColor = gray
			messageColor = gray
		case NOTICE:
			prefixColor = cyan
		case WARN:
			prefixColor = yellow
		case ERROR:
			prefixColor = red
		case FATAL:
			prefixColor = red
			messageColor = red
		}

		line = fmt.Sprintf("\x1b[%sm%s %-6s\x1b[0m \x1b[%sm%s\x1b[0m\n", prefixColor, now, level, messageColor, message)
	} else {
		line = fmt.Sprintf("%s %-6s %s\n", now, level, message)
	}

	// Make sure we're only outputing a line one at a time
	mutex.Lock()
	fmt.Fprint(l.Writer, line)
	mutex.Unlock()
}

var Discard = &Logger{
	Writer: ioutil.Discard,
}
