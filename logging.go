package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type Level int

const (
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "[DEBUG] "
	case LevelInfo:
		return "[INFO]  "
	case LevelWarn:
		return "[WARN]  "
	case LevelError:
		return "[ERROR] "
	default:
		return ""
	}
}

func (l *Level) Set(name string) error {
	switch strings.ToUpper(name) {
	case "DEBUG":
		*l = LevelDebug
	case "INFO":
		*l = LevelInfo
	case "WARN":
		*l = LevelWarn
	case "ERROR":
		*l = LevelError
	default:
		return errors.New("unknown name")
	}
	return nil
}

func logf(l Level, format string, args ...any) {
	if l < gLogLevel {
		return
	}
	log.Output(3, l.String()+fmt.Sprintf(format, args...))
}

func logp(l Level, args ...any) {
	if l < gLogLevel {
		return
	}
	log.Output(3, l.String()+fmt.Sprint(args...))
}

func debugf(f string, a ...any) { logf(LevelDebug, f, a...) }
func infof(f string, a ...any)  { logf(LevelInfo, f, a...) }
func warnf(f string, a ...any)  { logf(LevelWarn, f, a...) }
func errorf(f string, a ...any) { logf(LevelError, f, a...) }

// func debugp(a ...any) { logp(LevelDebug, a...) }
func infop(a ...any) { logp(LevelInfo, a...) }

// func warnp(a ...any)  { logp(LevelWarn, a...) }
func errorp(a ...any) { logp(LevelError, a...) }
