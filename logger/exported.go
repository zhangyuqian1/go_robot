package logger

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var (
	zaplogger *zap.Logger
	sugar     *zap.SugaredLogger // %s格式化输出用这个成员变量
)

// 这里定义时间，日志级别，堆栈等格式化
func getEncoder() zapcore.Encoder {
	config := zapcore.EncoderConfig{
		TimeKey:          "ts",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder, //日志级别大小表示
		EncodeTime:       zapcore.ISO8601TimeEncoder,  //时间格式
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder, //堆栈信息只需要打印到项目文件，不需要完整路径
		ConsoleSeparator: " ",
	}
	return zapcore.NewConsoleEncoder(config)
}

// 定义日志文件，控制台等
func getWriteSync(path string) zapcore.WriteSyncer {
	//file, _ := os.OpenFile("./log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	lumberJackLogger := &lumberjack.Logger{ //滚动输出
		Filename:   path,
		MaxSize:    1,  //单个文件最大M数，超过后自动滚动
		MaxBackups: 10, //滚动后最多保持的文件数
		MaxAge:     30, //保留的最大天数
		Compress:   false,
	}

	syncFile := zapcore.AddSync(lumberJackLogger) //文件输出
	syncConsole := zapcore.AddSync(os.Stdout)     //标准输出
	return zapcore.NewMultiWriteSyncer(syncConsole, syncFile)
	//zapcore.NewTee() 多个core
}

func Init(path string, l zapcore.Level) {

	config := getEncoder()
	write := getWriteSync(path)

	zaplogger = zap.New(
		zapcore.NewCore(config, write, l), //默认开启debug级别
		zap.AddCaller(),
		zap.AddStacktrace(zap.FatalLevel), //致命级问题，打印下详细堆栈
	)
	sugar = zaplogger.Sugar() //格式化句柄  类似fmt.printf("%s") 这种，有性能损耗

	zap.ReplaceGlobals(zaplogger)
	zap.RedirectStdLog(zaplogger) //自动替换掉系统的log
}

func Info(msg string, fields ...zap.Field) {
	zaplogger.Info(msg, fields...)
}

func Infof(template string, args ...interface{}) {
	sugar.Infof(template, args...)
}

// 错误日志
func Error(msg string, fields ...zap.Field) {
	zaplogger.Error(msg, fields...)
}

func Errorf(template string, args ...interface{}) {
	sugar.Errorf(template, args...)
}

// 告警，需要引起关注，但是不影响运行
func Warn(msg string, fields ...zap.Field) {
	zaplogger.Warn(msg, fields...)
}

func Warnf(template string, args ...interface{}) {
	sugar.Warnf(template, args...)
}

// 调试日志
func Debug(msg string, fields ...zap.Field) {
	zaplogger.Debug(msg, fields...)
}

func Debugf(template string, args ...interface{}) {
	sugar.Debugf(template, args...)
}

// 致命错误，调用后，程序会退出
func Fatal(msg string, fields ...zap.Field) {
	zaplogger.Fatal(msg, fields...)
}

func Fatalf(template string, args ...interface{}) {
	sugar.Fatalf(template, args...)
}
