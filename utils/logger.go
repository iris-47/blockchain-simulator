package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)

// Logger is a simple logger that supports different log levels
type Logger struct {
	logger *log.Logger
	prefix string
	level  int

	lock sync.Mutex
}

// init the default logger in case the user forgets to init it
func init() {
	LoggerInstance, _ = NewLogger("", INFO, "", true)
}

var LoggerInstance *Logger

// Craete a new logger instance
func NewLogger(logFile string, level int, prefix string, toStdout bool) (*Logger, error) {
	var file io.Writer
	var err error

	if logFile == "" {
		file = os.Stdout
	} else {
		file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		if toStdout {
			file = io.MultiWriter(file, os.Stdout)
		}
	}

	logger := log.New(file, "", log.Ltime)
	return &Logger{logger: logger, level: level, prefix: prefix}, nil
}

func (l *Logger) SetLevel(level int) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.level = level
}

// 获取调用者信息
func (l *Logger) getCallerInfo() string {
	// skip=3 是为了跳过 getCallerInfo 和当前的日志函数本身
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return ""
	}
	fileName := filepath.Base(file)
	return fmt.Sprintf("%s:%d", fileName, line)
}

func (l *Logger) log(level int, levelStr string, format string, v ...interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.level <= level {
		callerInfo := l.getCallerInfo()
		l.logger.SetPrefix(fmt.Sprintf("%s:[%s] %s ", l.prefix, levelStr, callerInfo))
		l.logger.Printf(format, v...)
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DEBUG, "DEBUG", format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.log(INFO, "INFO", format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(WARN, "WARN", format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ERROR, "ERROR", format, v...)
}
