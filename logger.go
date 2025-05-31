package gogi

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/dejaniskra/go-gi/internal/config"
)

type Field struct {
	Key   string
	Value any
}

var defaultLogger *Logger

type Logger struct {
	mu     sync.Mutex
	level  string
	format string
	out    io.Writer
}

func GetLogger() *Logger {
	if defaultLogger == nil {
		config := config.GetConfig()
		defaultLogger = newLogger(config.Log.Level, config.Log.Format)
	}
	return defaultLogger
}

func newLogger(level string, format string) *Logger {
	if defaultLogger != nil {
		return defaultLogger
	}
	return &Logger{level: level, format: format, out: os.Stdout}
}

func (l *Logger) log(level string, msg string) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.format == "JSON" {
		entry := map[string]any{
			"level":     level,
			"timestamp": timestamp,
			"message":   msg,
		}

		data, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(l.out, "Logger error: %v\n", err)
			return
		}
		fmt.Fprintln(l.out, string(data))
	} else {
		fmt.Fprintf(l.out, "[%s] %s: %s", timestamp, level, msg)
	}
}

func (l *Logger) debug(msg string) {
	l.log("DEBUG", msg)
}

func (l *Logger) info(msg string) {
	l.log("INFO", msg)
}

func (l *Logger) warn(msg string) {
	l.log("WARN", msg)
}

func (l *Logger) error(msg string) {
	l.log("ERROR", msg)
}

func (l *Logger) Debug(msg string) {
	l.debug(msg)
}

func (l *Logger) Info(msg string) {
	l.info(msg)
}

func (l *Logger) Warn(msg string) {
	l.warn(msg)
}

func (l *Logger) Error(msg string) {
	l.error(msg)
}
