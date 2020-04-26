package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// error logger
var zapLogger *zap.SugaredLogger
var zapLoggererr *zap.SugaredLogger
var zapLoggerclose *lumberjack.Logger
var zapLoggererrclose *lumberjack.Logger

var levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

// Close 关闭 日志的相关句柄
func Close() {
	zapLoggerclose.Close()
	zapLoggererrclose.Close()
}

// Init 日志 初始化
func Init(logConfigInfoName, logConfigErrorName, logLevel string) {
	level := getLoggerLevel(logLevel)
	zapLoggerclose = &lumberjack.Logger{
		Filename:   logConfigInfoName,
		MaxSize:    1 << 6, // megabytes 64MB
		LocalTime:  true,
		Compress:   true,
		MaxBackups: 30, // 最多保留30个备份
		MaxAge:     30, // days
	}
	syncInfoWriter := zapcore.AddSync(zapLoggerclose)

	zapLoggererrclose = &lumberjack.Logger{
		Filename:   logConfigErrorName,
		MaxSize:    1 << 6, // megabytes 64MB
		LocalTime:  true,
		Compress:   true,
		MaxBackups: 30, // 最多保留30个备份
		MaxAge:     30, // days
	}
	syncErrWriter := zapcore.AddSync(zapLoggererrclose)

	//syncWriter := os.Stdout
	encoder := zap.NewProductionEncoderConfig()
	//encoder := zap.NewDevelopmentEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder

	coreinfo := zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), syncInfoWriter, zap.NewAtomicLevelAt(level))
	coreerr := zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), syncErrWriter, zap.NewAtomicLevelAt(level))
	//core := zapcore.NewCore(zapcore.NewJSONEncoder(encoder), syncWriter, zap.NewAtomicLevelAt(level))
	// 函数行号多跳1级 这里封装了一层
	loggerinfo := zap.New(coreinfo, zap.AddCaller(), zap.AddCallerSkip(1))
	loggererr := zap.New(coreerr, zap.AddCaller(), zap.AddCallerSkip(1))
	zapLogger = loggerinfo.Sugar()
	zapLoggererr = loggererr.Sugar()
}

func Debug(args ...interface{}) {
	zapLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	zapLogger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	zapLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	zapLogger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	zapLoggererr.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	zapLoggererr.Warnf(template, args...)
}

func Error(args ...interface{}) {
	zapLoggererr.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	zapLoggererr.Errorf(template, args...)
}

// 以下会导致函数退出
func DPanic(args ...interface{}) {
	zapLoggererr.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	zapLoggererr.DPanicf(template, args...)
}

func Panic(args ...interface{}) {
	zapLoggererr.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	zapLoggererr.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	zapLoggererr.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	zapLoggererr.Fatalf(template, args...)
}

