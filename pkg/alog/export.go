package alog

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var _ ALogger = (*alog)(nil)

var Log ALogger

func init() {
	Log = New(context.TODO(), "json", "trace", false)
}

const (
	LevelTrace = slog.Level(-8)
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
	LevelFatal = slog.Level(12)
	LevelPanic = slog.Level(16)
)

type alog struct {
	*slog.Logger
	ctx context.Context
}

func toLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return LevelTrace
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	case "panic":
		return LevelPanic
	default:
		return LevelDebug
	}
}

func New(ctx context.Context, formatter string, level string, addSource bool) *alog {
	if ctx == nil {
		ctx = context.Background()
	}
	if formatter == "" {
		formatter = "json"
	}
	lv := toLevel(level)
	log := &alog{
		Logger: newLogger(formatter, lv, addSource),
		ctx:    ctx,
	}

	slog.SetDefault(log.Logger)
	return log
}

func (l *alog) WithContext(ctx context.Context) *alog {
	l.ctx = ctx
	return l
}

func (s *alog) enabled(level slog.Level) bool {
	return s.Logger.Enabled(s.ctx, level)
}

func (s *alog) handle(r slog.Record) error {
	return s.Logger.Handler().Handle(s.ctx, r)
}

func newLogger(logType string, level slog.Level, addSource bool) *slog.Logger {
	replace := func(groups []string, a slog.Attr) slog.Attr {
		// uncomment to remove the top-level time attribute
		// if a.Key == slog.TimeKey && len(groups) == 0 {
		// 	return slog.Attr{}
		// }

		// remove the directory from the source's filename
		if a.Key == slog.SourceKey && addSource {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}

		// customize the name of the level key and the output string
		// including custome level values.
		if a.Key == slog.LevelKey {
			// rename the level key from "level" to "lv"
			a.Key = "lv"

			// Handle custome level values
			lv := a.Value.Any().(slog.Level)
			// This could also look up the name from a map or other structure, but
			// this demonstrates using a switch statement to rename levels. For
			// maximum performance, the string values should be constants, but this
			// example uses the raw strings for readability.
			switch {
			case lv < LevelDebug:
				a.Value = slog.StringValue("TRACE")
			case lv < LevelInfo:
				a.Value = slog.StringValue("DEBUG")
			case lv < LevelWarn:
				a.Value = slog.StringValue("INFO")
			case lv < LevelError:
				a.Value = slog.StringValue("WARN")
			case lv < LevelFatal:
				a.Value = slog.StringValue("ERROR")
			case lv < LevelPanic:
				a.Value = slog.StringValue("FATAL")
			default:
				a.Value = slog.StringValue("PANIC")
			}
		}
		return a
	}

	if logType == "text" {
		tmp := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   addSource,
			Level:       level,
			ReplaceAttr: replace,
		}))
		slog.SetDefault(tmp)
		return tmp
	}

	tmp := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   addSource,
		Level:       slog.Level(level),
		ReplaceAttr: replace,
	}))
	slog.SetDefault(tmp)
	return tmp
}

func (s *alog) Trace(msg string, v ...any) {
	if !s.enabled(LevelTrace) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), LevelTrace, msg, pcs[0])
	r.Add(v...)
	_ = s.handle(r)
}

func (s *alog) Debug(msg string, v ...any) {
	if !s.enabled(LevelDebug) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), LevelDebug, msg, pcs[0])
	r.Add(v...)
	_ = s.handle(r)
}

func (s *alog) Info(msg string, v ...any) {
	if !s.enabled(LevelInfo) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, msg, pcs[0])
	r.Add(v...)
	_ = s.handle(r)
}

func (s *alog) Warn(msg string, v ...any) {
	if !s.enabled(LevelWarn) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelWarn, msg, pcs[0])
	r.Add(v...)
	_ = s.handle(r)
}

func (s *alog) Error(msg string, v ...any) {
	if !s.enabled(LevelError) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelError, msg, pcs[0])
	r.Add(v...)
	_ = s.handle(r)
}

func (s *alog) Fatal(msg string, v ...any) {
	if !s.enabled(LevelFatal) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), LevelFatal, msg, pcs[0])
	r.Add(v...)
	_ = s.handle(r)

	os.Exit(2)
}

func (s *alog) Panic(msg string, v ...any) {
	if !s.enabled(LevelPanic) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), LevelPanic, msg, pcs[0])
	r.Add(v...)
	_ = s.handle(r)

	panic(r.Message)
}

func (a *alog) With(args ...any) ALogger {
	// a.Logger = a.Logger.With(args...)
	return &alog{
		Logger: a.Logger.With(args...),
		ctx:    a.ctx,
	}
}

func (a *alog) WithGroup(name string) ALogger {
	// a.Logger = a.Logger.WithGroup(name)
	return &alog{
		Logger: a.Logger.WithGroup(name),
		ctx:    a.ctx,
	}
}

func Trace(msg string, v ...any) {
	log := Log.(*alog)
	if !log.enabled(LevelTrace) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), LevelTrace, msg, pcs[0])
	r.Add(v...)
	_ = log.handle(r)
}

func Debug(msg string, v ...any) {
	log := Log.(*alog)
	if !log.enabled(LevelDebug) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), LevelDebug, msg, pcs[0])
	r.Add(v...)
	_ = log.handle(r)
}

func Info(msg string, v ...any) {
	log := Log.(*alog)
	if !log.enabled(LevelInfo) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), LevelInfo, msg, pcs[0])
	r.Add(v...)
	_ = log.handle(r)
}

func Warn(msg string, v ...any) {
	log := Log.(*alog)
	if !log.enabled(LevelWarn) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), LevelWarn, msg, pcs[0])
	r.Add(v...)
	_ = log.handle(r)
}

func Error(msg string, v ...any) {
	log := Log.(*alog)
	if !log.enabled(LevelError) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), LevelError, msg, pcs[0])
	r.Add(v...)
	_ = log.handle(r)
}

func Fatal(msg string, v ...any) {
	log := Log.(*alog)
	if !log.enabled(LevelFatal) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), LevelFatal, msg, pcs[0])
	r.Add(v...)
	_ = log.handle(r)

	os.Exit(2)
}

func Panic(msg string, v ...any) {
	log := Log.(*alog)
	if !log.enabled(LevelPanic) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), LevelPanic, msg, pcs[0])
	r.Add(v...)
	_ = log.handle(r)

	panic(r.Message)
}
