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
	return [...]string{"debug", "info", "warn", "error"}[l]
}

func (l *Level) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "debug":
		*l = DEBUG
	case "info":
		*l = INFO
	case "warn":
		*l = WARN
	case "error":
		*l = ERROR
	default:
		return fmt.Errorf("invalid log level: %s", s)
	}

	return nil
}

type Format int

const (
	TEXT Format = iota
	JSON
)

func (f *Format) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "json":
		*f = JSON
	case "text":
		*f = TEXT
	default:
		return fmt.Errorf("invalid log format: %s", s)
	}

	return nil
}

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

func (l *Logger) log(level Level, msg string, fields ...Field) {
	l.logCtx(context.Background(), level, msg, fields...)
}

func (l *Logger) logCtx(ctx context.Context, level Level, msg string, fields ...Field) {
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

func (l *Logger) debug(msg string, fields ...Field) {
	l.log(DEBUG, msg, fields...)
}

func (l *Logger) info(msg string, fields ...Field) {
	l.log(INFO, msg, fields...)
}

func (l *Logger) warn(msg string, fields ...Field) {
	l.log(WARN, msg, fields...)
}

func (l *Logger) error(msg string, fields ...Field) {
	l.log(ERROR, msg, fields...)
}

func (l *Logger) debugCtx(ctx context.Context, msg string, fields ...Field) {
	l.logCtx(ctx, DEBUG, msg, fields...)
}

func (l *Logger) infoCtx(ctx context.Context, msg string, fields ...Field) {
	l.logCtx(ctx, INFO, msg, fields...)
}

func (l *Logger) warnCtx(ctx context.Context, msg string, fields ...Field) {
	l.logCtx(ctx, WARN, msg, fields...)
}

func (l *Logger) errorCtx(ctx context.Context, msg string, fields ...Field) {
	l.logCtx(ctx, ERROR, msg, fields...)
}

var defaultLogger *Logger

func InitGlobal(level Level, format Format) {
	defaultLogger = New(level, format)
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
