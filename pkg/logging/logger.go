package logging

import (
	"context"
	"fmt"
	"github.com/natefinch/lumberjack"
	"net/url"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const loggerKey = contextKey("logger")

var (
	// 生成默认日志器
	defaultLogger     *zap.SugaredLogger
	defaultLoggerOnce sync.Once
)

type lumberjackSink struct {
	*lumberjack.Logger
}

func (lumberjackSink) Sync() error {
	return nil
}

func NewLogger(debug bool) *zap.SugaredLogger {
	config := &zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Encoding:         encodingConsole,
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputFile,
		ErrorOutputPaths: outputFile,
	}

	ll := lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    64, //MB
		MaxBackups: 3,
		MaxAge:     30, //days
		Compress:   true,
	}

	zap.RegisterSink("lumberjack", func(*url.URL) (zap.Sink, error) {
		return lumberjackSink{
			Logger: &ll,
		}, nil
	})

	if debug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.Development = true
	}

	logger, err := config.Build()
	if err != nil {
		logger = zap.NewNop()
	}

	return logger.Sugar()
}

func DefaultLogger() *zap.SugaredLogger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = NewLogger(true)
	})
	return defaultLogger
}

// 绑定日志器到Context
func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

//返回Context中存在的日志器，不存在则创建新的日志器
func FromContext(ctx context.Context) *zap.SugaredLogger {
	if logger, ok := ctx.Value(loggerKey).(*zap.SugaredLogger); ok {
		return logger
	}
	return DefaultLogger()
}

const (
	timestamp  = "timestamp"
	severity   = "severity"
	logger     = "logger"
	caller     = "caller"
	message    = "message"
	stacktrace = "stacktrace"

	levelDebug     = "DEBUG"
	levelInfo      = "INFO"
	levelWarning   = "WARNING"
	levelError     = "ERROR"
	levelCritical  = "CRITICAL"
	levelAlert     = "ALERT"
	levelEmergency = "EMERGENCY"

	encodingJSON    = "json"
	encodingConsole = "console"
)

//var outputStderr = []string{"stderr"}
var logFile = "logs/app.log"
var outputFile = []string{"stderr", fmt.Sprintf("lumberjack:%s", logFile)}

var encoderConfig = zapcore.EncoderConfig{
	TimeKey:        timestamp,
	LevelKey:       severity,
	NameKey:        logger,
	CallerKey:      caller,
	MessageKey:     message,
	StacktraceKey:  stacktrace,
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    levelEncoder(),
	EncodeTime:     timeEncoder(),
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

func levelEncoder() zapcore.LevelEncoder {
	return func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		switch l {
		case zapcore.DebugLevel:
			enc.AppendString(levelDebug)
		case zapcore.InfoLevel:
			enc.AppendString(levelInfo)
		case zapcore.WarnLevel:
			enc.AppendString(levelWarning)
		case zapcore.ErrorLevel:
			enc.AppendString(levelError)
		case zapcore.DPanicLevel:
			enc.AppendString(levelCritical)
		case zapcore.PanicLevel:
			enc.AppendString(levelAlert)
		case zapcore.FatalLevel:
			enc.AppendString(levelEmergency)
		}
	}
}

// timeEncoder encodes the time as RFC3339 nano
func timeEncoder() zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339Nano))
	}
}

//-----------------模版方法------------------

func Debug(args ...interface{}) {
	DefaultLogger().Debug(args)
}

func Info(args ...interface{}) {
	DefaultLogger().Info(args)
}

func Warn(args ...interface{}) {
	DefaultLogger().Warn(args)
}

func Error(args ...interface{}) {
	DefaultLogger().Error(args)
}

func DPanic(args ...interface{}) {
	DefaultLogger().DPanic(args)
}

func Panic(args ...interface{}) {
	DefaultLogger().Panic(args)
}

func Fatal(args ...interface{}) {
	DefaultLogger().Fatal(args)
}

func Debugf(template string, args ...interface{}) {
	DefaultLogger().Debugf(template, args)
}

func Infof(template string, args ...interface{}) {
	DefaultLogger().Infof(template, args)
}

func Warnf(template string, args ...interface{}) {
	DefaultLogger().Warnf(template, args)
}

func Errorf(template string, args ...interface{}) {
	DefaultLogger().Errorf(template, args)
}

func DPanicf(template string, args ...interface{}) {
	DefaultLogger().DPanicf(template, args)
}

func Panicf(template string, args ...interface{}) {
	DefaultLogger().Panicf(template, args)
}

func Fatalf(template string, args ...interface{}) {
	DefaultLogger().Fatalf(template, args)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	DefaultLogger().Debugw(msg, keysAndValues)
}

func Infow(msg string, keysAndValues ...interface{}) {
	DefaultLogger().Infow(msg, keysAndValues)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	DefaultLogger().Warnw(msg, keysAndValues)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	DefaultLogger().Errorw(msg, keysAndValues)
}

func DPanicw(msg string, keysAndValues ...interface{}) {
	DefaultLogger().DPanicw(msg, keysAndValues)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	DefaultLogger().Panicw(msg, keysAndValues)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	DefaultLogger().Fatalw(msg, keysAndValues)
}
