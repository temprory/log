package main

import (
	"encoding/json"
	"fmt"
	tlog "github.com/temprory/log"
	"runtime"
	"strings"
)

var (
	// 按天切割日志文件，日志根目录下不设子目录，不限制单个日志文件大小
	fileWriter = &tlog.FileWriter{
		RootDir:     "./logs1/",       //日志根目录
		DirFormat:   "20060102-1504/", //日志根目录下无子目录
		FileFormat:  "20060102.log",   //日志文件命名规则，按天切割文件
		MaxFileSize: 1024 * 1024,      //单个日志文件最大size，0则不限制size
		EnableBufio: false,            //是否开启bufio
	}
)

type Writer struct{}

func (w *Writer) WriteLog(log *tlog.Log) (n int, err error) {
	value := log.Value

	_, file, line, ok := runtime.Caller(log.Depth + 1)
	if !ok {
		file = "???"
		line = -1
	} else {
		pos := strings.LastIndex(file, "/")
		if pos >= 0 {
			file = file[pos+1:]
		}
	}

	switch log.Level {
	case tlog.LEVEL_PRINT:

	case tlog.LEVEL_DEBUG:
		value = strings.Join([]string{log.Now.Format(log.Logger.Layout), fmt.Sprintf(" [Debug] [%s:%d] ", file, line), log.Value, "\n"}, "")
	case tlog.LEVEL_INFO:
		value = strings.Join([]string{log.Now.Format(log.Logger.Layout), fmt.Sprintf(" [ Info] [%s:%d] ", file, line), log.Value, "\n"}, "")
	case tlog.LEVEL_WARN:
		value = strings.Join([]string{log.Now.Format(log.Logger.Layout), fmt.Sprintf(" [ Warn] [%s:%d] ", file, line), log.Value, "\n"}, "")
	case tlog.LEVEL_ERROR:
		value = strings.Join([]string{log.Now.Format(log.Logger.Layout), fmt.Sprintf(" [Error] [%s:%d] ", file, line), log.Value, "\n"}, "")
	case tlog.LEVEL_PANIC:
		value = strings.Join([]string{log.Now.Format(log.Logger.Layout), fmt.Sprintf(" [Panic] [%s:%d] ", file, line), log.Value, "\n"}, "")
	case tlog.LEVEL_FATAL:
		value = strings.Join([]string{log.Now.Format(log.Logger.Layout), fmt.Sprintf(" [Fatal] [%s:%d] ", file, line), log.Value, "\n"}, "")
	default:
	}
	fmt.Println("--- log: ", log.File, log.Line, log.Depth, file, line)
	log.File = file
	log.Line = line

	data, err := json.Marshal(log)
	if err != nil {
		panic(err)
	}
	value = string(data)
	n, err = fileWriter.WriteString(value)
	fmt.Println(value)

	return n, err
}

func logInfo(data interface{}) {

}

func main() {
	tlog.SetOutput(nil)
	tlog.SetStructOutput(&Writer{})

	tlog.SetLevel(tlog.LEVEL_WARN)

	i := 0
	for {
		i++
		m := map[string]interface{}{
			"idx":   i,
			"key":   fmt.Sprintf("key %v", i),
			"value": fmt.Sprintf("value %v", i),
		}
		data, err := json.Marshal(&m)
		if err != nil {
			panic(err)
		}
		str := string(data)
		tlog.Debug(str)
		tlog.Info(str)
		tlog.Warn(str)
		tlog.Error(str)
	}
}
