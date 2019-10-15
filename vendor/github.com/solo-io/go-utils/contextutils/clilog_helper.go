package contextutils

import (
	"context"

	"github.com/solo-io/go-utils/clicore/constants"
)

/*
Long form:
	contextutils.LoggerFrom(ctx).Infow("message going to file only", zap.String("cli", "info that will go to the console and file")
	contextutils.LoggerFrom(ctx).Warnw("message going to file only", zap.String("cli", "a warning that will go to the console and file"))
	contextutils.LoggerFrom(ctx).Errorw("message going to file only", zap.String("cli", "an error that will go to the console and file")

Short form with the helper:
	contextutils.CliLogInfow(ctx, "this info log goes to file and console")
	contextutils.CliLogWarnw(ctx, "this warn log goes to file and console")
	contextutils.CliLogErrorw(ctx, "this error log goes to file and console")

Key-value pairs are also supported
- the extra key-value pairs get printed to the file log, not the console
	contextutils.CliLogInfow(ctx, "this message goes to the file and console", "extraKey", "extraValue")
*/

type cliLogLevel int

// Note that there is no Fatal log level. This is intentional.
// All errors should be surfaced up to the main entry point so that we can use
// Cobra's built-in error pipeline effectively.
const (
	cliLogLevelInfo cliLogLevel = iota + 1
	cliLogLevelWarn
	cliLogLevelError
)

func CliLogErrorw(ctx context.Context, message string, keysAndValues ...interface{}) {
	cliLogw(ctx, cliLogLevelError, message, constants.CliLoggerKey, keysAndValues...)
}
func CliLogWarnw(ctx context.Context, message string, keysAndValues ...interface{}) {
	cliLogw(ctx, cliLogLevelWarn, message, constants.CliLoggerKey, keysAndValues...)
}
func CliLogInfow(ctx context.Context, message string, keysAndValues ...interface{}) {
	cliLogw(ctx, cliLogLevelInfo, message, constants.CliLoggerKey, keysAndValues...)
}
func cliLogw(ctx context.Context, level cliLogLevel, message, cliLogKey string, keysAndValues ...interface{}) {
	log := LoggerFrom(ctx)
	kvs := []interface{}{cliLogKey, message}
	kvs = append(kvs, keysAndValues...)
	switch level {
	case cliLogLevelInfo:
		log.Infow(message, kvs...)
	case cliLogLevelWarn:
		log.Warnw(message, kvs...)
	case cliLogLevelError:
		log.Errorw(message, kvs...)
	}
}
