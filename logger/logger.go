package logger

import (
	"Hexo-AutoCD/config"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

var (
	// Log 全局日志实例
	Log *logrus.Logger
)

// CustomFormatter 自定义的日志格式化器
type CustomFormatter struct {
	TimestampFormat string
	DisableColors   bool
}

// Format 实现logrus.Formatter接口
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	// 时间格式化 (蓝色)
	timestamp := entry.Time.Format(f.TimestampFormat)
	if !f.DisableColors {
		fmt.Fprintf(b, "\x1b[34m[%s]\x1b[0m ", timestamp)
	} else {
		fmt.Fprintf(b, "[%s] ", timestamp)
	}

	// 日志级别 (不同颜色)
	var levelColor string
	if !f.DisableColors {
		switch entry.Level {
		case logrus.DebugLevel:
			levelColor = "\x1b[33m" // 黄色
		case logrus.InfoLevel:
			levelColor = "\x1b[32m" // 绿色
		case logrus.WarnLevel:
			levelColor = "\x1b[35m" // 紫色
		case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
			levelColor = "\x1b[31m" // 红色
		default:
			levelColor = "\x1b[36m" // 青色
		}
		fmt.Fprintf(b, "%s[%s]\x1b[0m ", levelColor, strings.ToUpper(entry.Level.String()))
	} else {
		fmt.Fprintf(b, "[%s] ", strings.ToUpper(entry.Level.String()))
	}

	// 消息
	fmt.Fprintf(b, "%s", entry.Message)

	// 添加字段信息，格式化输出
	if len(entry.Data) > 0 {
		// 先处理函数和文件信息（如果有的话）
		if funcName, ok := entry.Data["函数"]; ok {
			if fileInfo, ok2 := entry.Data["文件"]; ok2 {
				if !f.DisableColors {
					fmt.Fprintf(b, " \x1b[36m[%s@%v]\x1b[0m", funcName, fileInfo)
				} else {
					fmt.Fprintf(b, " [%s@%v]", funcName, fileInfo)
				}
				// 从Data中删除这两个字段，避免重复输出
				delete(entry.Data, "函数")
				delete(entry.Data, "文件")
			}
		}

		// 处理其它字段
		if len(entry.Data) > 0 {
			fieldStr := ""
			for k, v := range entry.Data {
				if k != "time" && k != "level" && k != "msg" {
					if fieldStr != "" {
						fieldStr += ", "
					}
					fieldStr += fmt.Sprintf("%s=%v", k, v)
				}
			}

			if fieldStr != "" {
				if !f.DisableColors {
					fmt.Fprintf(b, " \x1b[90m{%s}\x1b[0m", fieldStr) // 灰色显示其他字段
				} else {
					fmt.Fprintf(b, " {%s}", fieldStr)
				}
			}
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

// Init 初始化日志系统
func Init() error {
	// 创建日志目录
	logDir := filepath.Dir(config.Config.Logs.Path)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 创建日志实例
	Log = logrus.New()

	// 配置日志级别
	level, err := logrus.ParseLevel(config.Config.Logs.Level)
	if err != nil {
		level = logrus.InfoLevel // 默认使用 info 级别
	}
	Log.SetLevel(level)

	// 添加调用者信息的钩子
	Log.AddHook(&CallerHook{})

	// 根据配置选择格式化器
	var consoleFormatter, fileFormatter logrus.Formatter

	if strings.ToLower(config.Config.Logs.Format) == "json" {
		// JSON格式
		jsonFormatter := &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "时间",
				logrus.FieldKeyLevel: "级别",
				logrus.FieldKeyMsg:   "消息",
			},
		}
		consoleFormatter = jsonFormatter
		fileFormatter = jsonFormatter
	} else {
		// 文本格式
		consoleFormatter = &CustomFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			DisableColors:   false, // 控制台输出使用颜色
		}
		fileFormatter = &CustomFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			DisableColors:   true, // 文件输出不使用颜色
		}
	}

	// 设置控制台输出配置
	Log.SetFormatter(consoleFormatter)

	// 创建文件输出钩子
	fileHook := NewFileHook(
		config.Config.Logs.Path,
		config.Config.Logs.MaxSize,
		config.Config.Logs.MaxBackups,
		config.Config.Logs.MaxAge,
		fileFormatter)
	Log.AddHook(fileHook)

	// 记录初始化成功日志
	Log.Info("日志系统初始化成功")
	return nil
}

// CallerHook 是一个自定义的logrus钩子，用于添加调用者信息
type CallerHook struct{}

// Levels 指定钩子适用于哪些日志级别
func (hook *CallerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 在日志事件发生时添加调用者信息
func (hook *CallerHook) Fire(entry *logrus.Entry) error {
	// 获取调用者信息，跳过日志库本身的调用
	if pc, file, line, ok := runtime.Caller(6); ok {
		funcName := runtime.FuncForPC(pc).Name()
		// 移除包前缀，只保留函数名
		if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
			funcName = funcName[lastSlash+1:]
		}
		// 移除函数的包名
		if lastDot := strings.LastIndex(funcName, "."); lastDot >= 0 {
			funcName = funcName[lastDot+1:]
		}
		// 获取短文件名
		if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
			file = file[lastSlash+1:]
		}
		// 添加到日志字段
		entry.Data["函数"] = funcName
		entry.Data["文件"] = fmt.Sprintf("%s:%d", file, line)
	}
	return nil
}

// WithField 返回带有指定字段的日志实例
func WithField(key string, value interface{}) *logrus.Entry {
	return Log.WithField(key, value)
}

// WithFields 返回带有指定多个字段的日志实例
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}

// WithError 返回带有错误信息的日志实例
func WithError(err error) *logrus.Entry {
	return Log.WithError(err)
}

// 以下是各种日志级别的便捷函数

// Trace 记录跟踪级别日志
func Trace(args ...interface{}) {
	Log.Trace(args...)
}

// Tracef 记录格式化的跟踪级别日志
func Tracef(format string, args ...interface{}) {
	Log.Tracef(format, args...)
}

// Debug 记录调试级别日志
func Debug(args ...interface{}) {
	Log.Debug(args...)
}

// Debugf 记录格式化的调试级别日志
func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

// Info 记录信息级别日志
func Info(args ...interface{}) {
	Log.Info(args...)
}

// Infof 记录格式化的信息级别日志
func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

// Warn 记录警告级别日志
func Warn(args ...interface{}) {
	Log.Warn(args...)
}

// Warnf 记录格式化的警告级别日志
func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
}

// Error 记录错误级别日志
func Error(args ...interface{}) {
	Log.Error(args...)
}

// Errorf 记录格式化的错误级别日志
func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

// Fatal 记录严重错误级别日志并结束程序
func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

// Fatalf 记录格式化的严重错误级别日志并结束程序
func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

// Panic 记录重大错误级别日志并触发 panic
func Panic(args ...interface{}) {
	Log.Panic(args...)
}

// Panicf 记录格式化的重大错误级别日志并触发 panic
func Panicf(format string, args ...interface{}) {
	Log.Panicf(format, args...)
}

// FileHook 文件日志钩子
type FileHook struct {
	writer    io.Writer
	formatter logrus.Formatter
}

// NewFileHook 创建新的文件日志钩子
func NewFileHook(filePath string, maxSize, maxBackups, maxAge int, formatter logrus.Formatter) *FileHook {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize,    // 每个日志文件最大尺寸，单位是MB
		MaxBackups: maxBackups, // 保留的旧日志文件最大数量
		MaxAge:     maxAge,     // 保留的旧日志文件最大天数
		Compress:   true,       // 是否压缩旧日志文件
	}

	return &FileHook{
		writer:    lumberjackLogger,
		formatter: formatter,
	}
}

// Levels 返回所有日志级别
func (h *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 将日志写入文件
func (h *FileHook) Fire(entry *logrus.Entry) error {
	line, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = h.writer.Write(line)
	return err
}
