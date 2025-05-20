package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR"}[l]
}

type Format int

const (
	TEXT Format = iota
	JSON
)

type Field struct {
	Key   string
	Value any
}

type contextKey string

const logContextKey contextKey = "logger_fields"

func WithFields(ctx context.Context, fields ...Field) context.Context {
	existing, _ := ctx.Value(logContextKey).([]Field)
	return context.WithValue(ctx, logContextKey, append(existing, fields...))
}

func FieldsFromContext(ctx context.Context) []Field {
	fields, _ := ctx.Value(logContextKey).([]Field)
	return fields
}

type Logger struct {
	mu     sync.Mutex
	level  Level
	format Format
	out    io.Writer
}

func New(level Level, format Format) *Logger {
	return &Logger{level: level, format: format, out: os.Stdout}
}

func (l *Logger) Log(level Level, msg string, fields ...Field) {
	l.LogCtx(context.Background(), level, msg, fields...)
}

func (l *Logger) LogCtx(ctx context.Context, level Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	allFields := append(FieldsFromContext(ctx), fields...)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.format == JSON {
		entry := map[string]any{
			"level":     level.String(),
			"timestamp": timestamp,
			"message":   msg,
		}
		for _, f := range allFields {
			entry[f.Key] = f.Value
		}

		data, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(l.out, "Logger error: %v\n", err)
			return
		}
		fmt.Fprintln(l.out, string(data))

	} else {
		fmt.Fprintf(l.out, "[%s] %s: %s", timestamp, level.String(), msg)
		for _, f := range allFields {
			fmt.Fprintf(l.out, " %s=%v", f.Key, f.Value)
		}
		fmt.Fprintln(l.out)
	}
}

func (l *Logger) Debug(msg string, fields ...Field) {
	l.Log(DEBUG, msg, fields...)
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.Log(INFO, msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.Log(WARN, msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.Log(ERROR, msg, fields...)
}

func (l *Logger) DebugCtx(ctx context.Context, msg string, fields ...Field) {
	l.LogCtx(ctx, DEBUG, msg, fields...)
}

func (l *Logger) InfoCtx(ctx context.Context, msg string, fields ...Field) {
	l.LogCtx(ctx, INFO, msg, fields...)
}

func (l *Logger) WarnCtx(ctx context.Context, msg string, fields ...Field) {
	l.LogCtx(ctx, WARN, msg, fields...)
}

func (l *Logger) ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	l.LogCtx(ctx, ERROR, msg, fields...)
}

var defaultLogger *Logger

func InitGlobal(level Level, format Format) {
	defaultLogger = New(level, format)
}

func Debug(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, fields...)
	}
}

func DebugCtx(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.DebugCtx(ctx, msg, fields...)
	}
}

func InfoCtx(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.InfoCtx(ctx, msg, fields...)
	}
}

func WarnCtx(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.WarnCtx(ctx, msg, fields...)
	}
}

func ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.ErrorCtx(ctx, msg, fields...)
	}
}
