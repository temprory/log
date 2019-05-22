package log

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	LEVEL_PRINT = iota
	LEVEL_DEBUG
	LEVEL_INFO
	LEVEL_WARN
	LEVEL_ERROR
	LEVEL_PANIC
	LEVEL_FATAL
	LEVEL_NONE
)

var (
	BuildDir = ""

	logsep   = ""
	inittime = time.Now()

	DefaultLogLevel      = LEVEL_DEBUG
	DefaultLogDepth      = 2
	DefaultLogWriter     = os.Stdout
	DefaultLogTimeLayout = "2006-01-02 15:04:05.000"

	filepaths = []string{}

	DefaultLogger = NewLogger()
)

func init() {

	gopath := os.Getenv("GOPATH")
	if len(gopath) > 0 {
		if runtime.GOOS == "windows" {
			arr := strings.Split(gopath, ";")
			if len(arr) > 1 {
				filepaths = append(filepaths, arr...)
			} else {
				filepaths = append(filepaths, gopath)
			}
		} else {
			arr := strings.Split(gopath, ":")
			if len(arr) > 1 {
				filepaths = append(filepaths, arr...)
			} else {
				filepaths = append(filepaths, gopath)
			}
		}
	}

	goroot := os.Getenv("GOROOT")
	if len(gopath) > 0 {
		filepaths = append(filepaths, goroot)
	}

	for i, v := range filepaths {
		filepaths[i] = strings.Replace(v, `\`, `/`, -1)
		if strings.HasSuffix(filepaths[i], "/") {
			filepaths[i] = filepaths[i][:len(filepaths[i])-1]
		}
		filepaths[i] += `/src/`
	}

	wd, err := os.Getwd()
	if err == nil {
		wd = strings.Replace(wd, `\`, `/`, -1)
		if strings.HasSuffix(wd, "/") {
			wd = wd[:len(wd)-1]
		}
		pos := strings.LastIndex(wd, "/")
		if pos > 0 {
			wd = wd[:pos+1]
			filepaths = append([]string{wd}, filepaths...)
		}
	}

	if BuildDir != "" {
		BuildDir = strings.Replace(BuildDir, `\`, `/`, -1)
		if strings.HasSuffix(BuildDir, "/") {
			BuildDir = BuildDir[:len(BuildDir)-1]
		}
		pos := strings.LastIndex(BuildDir, "/")
		if pos > 0 {
			BuildDir = BuildDir[:pos+1]
			filepaths = append([]string{BuildDir}, filepaths...)
		}
	}

	// fmt.Println("--- filepaths:", filepaths)

	DefaultLogger.depth = DefaultLogDepth + 1
}

func LevelText(lvl int) string {
	switch lvl {
	case LEVEL_PRINT:
		return "Print"
	case LEVEL_DEBUG:
		return "Debug"
	case LEVEL_INFO:
		return "Info"
	case LEVEL_WARN:
		return "Warn"
	case LEVEL_ERROR:
		return "Error"
	case LEVEL_PANIC:
		return "Panic"
	case LEVEL_FATAL:
		return "Fatal"
	case LEVEL_NONE:
	default:
	}
	return "Unknown LVL"
}

type Log struct {
	Now    time.Time `json:"Now"`
	Depth  int       `json:"Depth"`
	Level  int       `json:"Level"`
	Line   int       `json:"Line"`
	File   string    `json:"File"`
	Value  string    `json:"Value"`
	Logger *Logger   `json:"-"`
}

type ILogWriter interface {
	WriteLog(log *Log) (n int, err error)
}

type LogWriter struct {
	writers []ILogWriter
}

func (w *LogWriter) WriteLog(log *Log) (n int, err error) {
	for _, v := range w.writers {
		v.WriteLog(log)
	}
	return 0, nil
}

func MultiLogWriter(writers ...ILogWriter) ILogWriter {
	w := &LogWriter{}
	for _, v := range writers {
		w.writers = append(w.writers, v.(ILogWriter))
	}
	return w
}

type Logger struct {
	sync.Mutex
	Writer    io.Writer
	LogWriter ILogWriter
	depth     int
	Level     int
	Layout    string
	Formater  func(log *Log) string
	FullPath  bool
	// filepaths []string
}

// func (logger *Logger) AddFileIgnorePath(path string) {
// 	path = strings.Replace(path, `\`, `/`, -1)
// 	for strings.HasPrefix(path, "/") {
// 		path = path[1:]
// 	}
// 	for strings.HasSuffix(path, "/") {
// 		path = path[:len(path)-1]
// 	}
// 	if len(path) > 0 {
// 		path += "/"
// 	}
// 	for i, v := range logger.filepaths {
// 		logger.filepaths[i] += v
// 	}
// }

func (logger *Logger) Printf(format string, v ...interface{}) {
	logger.Lock()
	if logger.Writer != nil {
		fmt.Fprintf(logger.Writer, fmt.Sprintf(format, v...))
	}
	if logger.LogWriter != nil {
		log := &Log{
			Depth:  logger.depth,
			Level:  LEVEL_PRINT,
			Value:  fmt.Sprintf(format, v...),
			Logger: logger,
		}
		logger.LogWriter.WriteLog(log)
	}
	logger.Unlock()
}

func (logger *Logger) Println(v ...interface{}) {
	logger.Lock()
	if logger.Writer != nil {
		fmt.Fprintln(logger.Writer, v...)
	}
	if logger.LogWriter != nil {
		log := &Log{
			Depth:  logger.depth,
			Level:  LEVEL_PRINT,
			Value:  fmt.Sprintln(v...),
			Logger: logger,
		}
		logger.LogWriter.WriteLog(log)
	}
	logger.Unlock()
}

func (logger *Logger) Debug(format string, v ...interface{}) {
	if LEVEL_DEBUG >= logger.Level {
		logger.Lock()
		now := time.Now()
		log := &Log{
			Now:    now,
			Depth:  logger.depth,
			Level:  LEVEL_DEBUG,
			Value:  fmt.Sprintf(format, v...),
			Logger: logger,
		}
		if logger.Writer != nil {
			fmt.Fprintln(logger.Writer, logger.Formater(log))
		}
		logger.Unlock()
		if logger.LogWriter != nil {
			logger.LogWriter.WriteLog(log)
		}
	}
}

func (logger *Logger) Info(format string, v ...interface{}) {
	if LEVEL_INFO >= logger.Level {
		logger.Lock()
		now := time.Now()
		log := &Log{
			Now:    now,
			Depth:  logger.depth,
			Level:  LEVEL_INFO,
			Value:  fmt.Sprintf(format, v...),
			Logger: logger,
		}
		if logger.Writer != nil {
			fmt.Fprintln(logger.Writer, logger.Formater(log))
		}
		logger.Unlock()
		if logger.LogWriter != nil {
			logger.LogWriter.WriteLog(log)
		}
	}
}

func (logger *Logger) Warn(format string, v ...interface{}) {
	if LEVEL_WARN >= logger.Level {
		logger.Lock()
		now := time.Now()
		log := &Log{
			Now:    now,
			Depth:  logger.depth,
			Level:  LEVEL_WARN,
			Value:  fmt.Sprintf(format, v...),
			Logger: logger,
		}
		if logger.Writer != nil {
			fmt.Fprintln(logger.Writer, logger.Formater(log))
		}
		logger.Unlock()
		if logger.LogWriter != nil {
			logger.LogWriter.WriteLog(log)
		}
	}
}

func (logger *Logger) Error(format string, v ...interface{}) {
	if LEVEL_ERROR >= logger.Level {
		logger.Lock()
		now := time.Now()
		log := &Log{
			Now:    now,
			Depth:  logger.depth,
			Level:  LEVEL_ERROR,
			Value:  fmt.Sprintf(format, v...),
			Logger: logger,
		}
		if logger.Writer != nil {
			fmt.Fprintln(logger.Writer, logger.Formater(log))
		}
		logger.Unlock()
		if logger.LogWriter != nil {
			logger.LogWriter.WriteLog(log)
		}
	}
}

func (logger *Logger) Panic(format string, v ...interface{}) {
	if LEVEL_PANIC >= logger.Level {
		logger.Lock()
		now := time.Now()
		log := &Log{
			Now:    now,
			Depth:  logger.depth,
			Level:  LEVEL_PANIC,
			Value:  fmt.Sprintf(format, v...),
			Logger: logger,
		}
		s := logger.Formater(log)
		if logger.Writer != nil {
			fmt.Fprintln(logger.Writer, s)
		}
		logger.Unlock()
		if logger.LogWriter != nil {
			logger.LogWriter.WriteLog(log)
		}
		panic(errors.New(s))
	}
}

func (logger *Logger) Fatal(format string, v ...interface{}) {
	if LEVEL_FATAL >= logger.Level {
		logger.Lock()
		now := time.Now()
		log := &Log{
			Now:    now,
			Depth:  logger.depth,
			Level:  LEVEL_FATAL,
			Value:  fmt.Sprintf(format, v...),
			Logger: logger,
		}
		if logger.Writer != nil {
			fmt.Fprintln(logger.Writer, logger.Formater(log))
		}
		logger.Unlock()
		if logger.LogWriter != nil {
			logger.LogWriter.WriteLog(log)
		}
		os.Exit(-1)
	}
}

func (logger *Logger) SetLevel(level int) {
	if level >= 0 && level <= LEVEL_NONE {
		logger.Level = level
	} else {
		log.Fatal(fmt.Errorf("log SetLogLevel Error: Invalid Level - %d\n", level))
	}
}

func (logger *Logger) SetOutput(out io.Writer) {
	logger.Writer = out
}

func (logger *Logger) SetStructOutput(out ILogWriter) {
	logger.LogWriter = out
}

func (logger *Logger) SetFormater(f func(log *Log) string) {
	logger.Formater = f
}

func (logger *Logger) defaultLogFormater(log *Log) string {
	_, file, line, ok := runtime.Caller(log.Depth)
	if !ok {
		file = "???"
		line = -1
	} else {
		if logger.FullPath {
			for _, v := range filepaths {
				tmp := strings.Replace(file, v, "", 1)
				if tmp != file {
					file = tmp
					break
				}
			}
		} else {
			pos := strings.LastIndex(file, "/")
			if pos >= 0 {
				file = file[pos+1:]
			}
		}
	}

	log.File = file
	log.Line = line
	switch log.Level {
	case LEVEL_DEBUG:
		return strings.Join([]string{log.Now.Format(logger.Layout), fmt.Sprintf(" [Debug] [%s:%d] ", file, line), log.Value}, "")
	case LEVEL_INFO:
		return strings.Join([]string{log.Now.Format(logger.Layout), fmt.Sprintf(" [ Info] [%s:%d] ", file, line), log.Value}, "")
	case LEVEL_WARN:
		return strings.Join([]string{log.Now.Format(logger.Layout), fmt.Sprintf(" [ Warn] [%s:%d] ", file, line), log.Value}, "")
	case LEVEL_ERROR:
		return strings.Join([]string{log.Now.Format(logger.Layout), fmt.Sprintf(" [Error] [%s:%d] ", file, line), log.Value}, "")
	case LEVEL_PANIC:
		return strings.Join([]string{log.Now.Format(logger.Layout), fmt.Sprintf(" [Panic] [%s:%d] ", file, line), log.Value}, "")
	case LEVEL_FATAL:
		return strings.Join([]string{log.Now.Format(logger.Layout), fmt.Sprintf(" [Fatal] [%s:%d] ", file, line), log.Value}, "")
	default:
	}
	return ""
}

func (logger *Logger) SetLogTimeFormat(layout string) {
	logger.Layout = layout
}

/********* default logger *********/
func Printf(fmtstr string, v ...interface{}) {
	DefaultLogger.Printf(fmtstr, v...)
}

func Println(v ...interface{}) {
	DefaultLogger.Println(v...)
}

func Debug(format string, v ...interface{}) {
	DefaultLogger.Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	DefaultLogger.Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	DefaultLogger.Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	DefaultLogger.Error(format, v...)
}

func Panic(format string, v ...interface{}) {
	DefaultLogger.Panic(format, v...)
}

func Fatal(format string, v ...interface{}) {
	DefaultLogger.Fatal(format, v...)
}

func SetLevel(level int) {
	DefaultLogger.SetLevel(level)
}

func SetOutput(out io.Writer) {
	DefaultLogger.SetOutput(out)
}

func SetStructOutput(out ILogWriter) {
	DefaultLogger.SetStructOutput(out)
}

func SetFormater(f func(log *Log) string) {
	DefaultLogger.SetFormater(f)
}

func SetLogTimeFormat(layout string) {
	DefaultLogger.SetLogTimeFormat(layout)
}

func SetBuildDir(dir string) {
	if dir != "" {
		BuildDir = dir
		if strings.HasSuffix(BuildDir, "/") {
			BuildDir = BuildDir[:len(BuildDir)-1]
		}
		pos := strings.LastIndex(BuildDir, "/")
		if pos > 0 {
			BuildDir = BuildDir[:pos+1]
			filepaths = append([]string{BuildDir}, filepaths...)
		}
	}
}

func LogWithFormater(lvl int, depth int, layout string, format string, v ...interface{}) string {
	now := time.Now()
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = -1
	} else {
		pos := strings.LastIndex(file, "/")
		if pos >= 0 {
			file = file[pos+1:]
		}
	}

	switch lvl {
	case LEVEL_DEBUG:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [Debug] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_INFO:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [ Info] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_WARN:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [ Warn] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_ERROR:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [Error] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_PANIC:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [Panic] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_FATAL:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [Fatal] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	default:
	}
	return ""
}

func NewLogger() *Logger {
	logger := &Logger{
		Level:    DefaultLogLevel,
		depth:    DefaultLogDepth,
		Writer:   DefaultLogWriter,
		Layout:   DefaultLogTimeLayout,
		FullPath: BuildDir != "",
		//filepaths: append([]string{}, filepaths...),
	}
	logger.Formater = logger.defaultLogFormater
	return logger
}
