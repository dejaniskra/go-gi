package gogi

import (
	"context"
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
		defaultLogger = New(config.Log.Level, config.Log.Format)
	}
	return defaultLogger
}

func New(level string, format string) *Logger {
	if defaultLogger != nil {
		return defaultLogger
	}
	return &Logger{level: level, format: format, out: os.Stdout}
}

func (l *Logger) log(level string, msg string, fields ...Field) {
	l.logCtx(context.Background(), level, msg, fields...)
}

func (l *Logger) logCtx(ctx context.Context, level string, msg string, fields ...Field) {
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

func (l *Logger) debug(msg string, fields ...Field) {
	l.log("DEBUG", msg, fields...)
}

func (l *Logger) info(msg string, fields ...Field) {
	l.log("INFO", msg, fields...)
}

func (l *Logger) warn(msg string, fields ...Field) {
	l.log("WARN", msg, fields...)
}

func (l *Logger) error(msg string, fields ...Field) {
	l.log("ERROR", msg, fields...)
}

func (l *Logger) debugCtx(ctx context.Context, msg string, fields ...Field) {
	l.logCtx(ctx, "DEBUG", msg, fields...)
}

func (l *Logger) infoCtx(ctx context.Context, msg string, fields ...Field) {
	l.logCtx(ctx, "INFO", msg, fields...)
}

func (l *Logger) warnCtx(ctx context.Context, msg string, fields ...Field) {
	l.logCtx(ctx, "WARN", msg, fields...)
}

func (l *Logger) errorCtx(ctx context.Context, msg string, fields ...Field) {
	l.logCtx(ctx, "ERROR", msg, fields...)
}

func Debug(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.debug(msg, fields...)
	}
}

func Info(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.info(msg, fields...)
	}
}

func Warn(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.warn(msg, fields...)
	}
}

func Error(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.error(msg, fields...)
	}
}

func DebugCtx(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.debugCtx(ctx, msg, fields...)
	}
}

func InfoCtx(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.infoCtx(ctx, msg, fields...)
	}
}

func WarnCtx(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.warnCtx(ctx, msg, fields...)
	}
}

func ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.errorCtx(ctx, msg, fields...)
	}
}
