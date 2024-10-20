package utils

import (
	"BlockChainSimulator/config"
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

// must be created before using the logger
var LoggerInstance *Logger

// Craete a new logger instance
func NewLogger(args *config.Args, level string, toStdout bool, toFile bool) (*Logger, error) {
	var file io.Writer
	var err error

	// check if the log path exists, if not, create it
	if _, err := os.Stat(config.LogPath); os.IsNotExist(err) {
		os.Mkdir(config.LogPath, os.ModePerm)
	}

	// set the prefix and log file name according to the args
	prefix := fmt.Sprintf("[S%dN%d]", args.ShardID, args.NodeID)
	logFile := fmt.Sprintf("%sS%dN%d.log", config.LogPath, args.ShardID, args.NodeID)
	if args.IsClient {
		prefix = "[Client]"
		logFile = fmt.Sprintf("%sclient.log", config.LogPath)
	}

	if !toFile {
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
	return &Logger{logger: logger, level: str2Level(level), prefix: prefix}, nil
}

func (l *Logger) SetPrefix(prefix string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.prefix = prefix
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

func str2Level(level string) int {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
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
