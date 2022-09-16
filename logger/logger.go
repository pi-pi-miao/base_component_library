package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

var (
	SugaredLogger *zap.SugaredLogger
	coreLevel     = map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"error": zapcore.ErrorLevel,
	}
)

type LogConfig struct {
	LogLevel string
	LogFile  string
	IsDebug  string
	logType  string
}

func ZnTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000000"))
}

func InitLog(filePath string, level string, maxSize int, maxBackups int, maxAge int) {
	var (
		l zapcore.Level
		w zapcore.WriteSyncer
	)
	l = zapcore.DebugLevel
	if v, ok := coreLevel[level]; ok {
		l = v
	}
	//w = zapcore.AddSync(NewRotate(maxBackups,maxSize,filePath))
	w = zapcore.AddSync(NewAsync(maxBackups, maxSize, filePath))
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = ZnTimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		w,
		zap.NewAtomicLevelAt(l),
	)
	SugaredLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel)).Sugar()
	return
}

func Debug(args ...interface{}) {
	SugaredLogger.Debug(args...)
}

func Info(args ...interface{}) {
	SugaredLogger.Info(args...)
}

func Error(args ...interface{}) {
	SugaredLogger.Error(args...)
}

func Debugf(format string, args ...interface{}) {
	SugaredLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	SugaredLogger.Infof(format, args...)
}

func Errorf(format string, args ...interface{}) {
	SugaredLogger.Errorf(format, args...)
}

type LogObject struct {
	l *zap.SugaredLogger
}

// InitLogJsonQoeBilling ：以对象的方式运行log，返回一个日志对象
func InitLogJsonQoeBilling(filePath, name string, level string, maxBackups int) *LogObject {
	var (
		l zapcore.Level
	)
	l = zapcore.DebugLevel
	if v, ok := coreLevel[level]; ok {
		l = v
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			MessageKey: "msg",
		}),
		zapcore.AddSync(NewQoeBilling(filePath, name, maxBackups)),
		zap.NewAtomicLevelAt(l),
	)
	obj := &LogObject{}
	obj.l = zap.New(core).Sugar()
	return obj
}

func (l *LogObject) Debug(args ...interface{}) {
	l.l.Debug(args...)
}

func (l *LogObject) Info(args ...interface{}) {
	l.l.Info(args...)
}

func (l *LogObject) Error(args ...interface{}) {
	l.l.Error(args...)
}

func (l *LogObject) Debugf(format string, args ...interface{}) {
	l.l.Debugf(format, args...)
}

func (l *LogObject) Infof(format string, args ...interface{}) {
	l.l.Infof(format, args...)
}

func (l *LogObject) Errorf(format string, args ...interface{}) {
	l.l.Errorf(format, args...)
}
