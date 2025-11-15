package logger

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

type logger struct {
	z zerolog.Logger
}

func New() Logger {
	return &logger{z: zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.DateTime}).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		CallerWithSkipFrameCount(3).
		Logger(),
	}
}

func (l *logger) Printf(format string, a ...interface{}) {
	l.z.Printf(format, a...)
}

func (l *logger) Error(a ...interface{}) {
	l.z.Error().Msgf("%v", a)
}

func (l *logger) Errorf(format string, a ...interface{}) {
	l.z.Error().Msgf(format, a...)
}

func (l *logger) Fatalf(format string, a ...interface{}) {
	l.z.Fatal().Msgf(format, a...)
}

func (l *logger) Warnf(format string, a ...interface{}) {
	l.z.Warn().Msgf(format, a...)
}

func (l *logger) Infof(format string, a ...interface{}) {
	l.z.Info().Msgf(format, a...)
}
