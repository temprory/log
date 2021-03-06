package main

import (
	"fmt"
	"github.com/temprory/log"
	"io"
	"os"
)

func main() {
	// 按天切割日志文件，日志根目录下子目录按天存储，并限制单个日志文件大小
	fileWriter := &log.FileWriter{
		RootDir:     "./logs2/",       //日志根目录
		DirFormat:   "200601021504/",  //日志根目录下按天分割子目录
		FileFormat:  "20060102.log",   //日志文件命名规则，按天切割文件
		MaxFileSize: 1024 * 1024 * 32, //日志文件最大size，按size切割日志文件
		EnableBufio: false,            //是否启用bufio
	}
	out := io.MultiWriter(os.Stdout)
	log.SetOutput(out)

	// writer := log.MultiLogWriter(fileWriter)
	// log.SetStructOutput(writer)
	log.SetStructOutput(fileWriter)

	log.SetLevel(log.LEVEL_WARN)

	i := 0
	for {
		i++
		log.Debug(fmt.Sprintf("log %d", i))
		log.Info(fmt.Sprintf("log %d", i))
		log.Warn(fmt.Sprintf("log %d", i))
		log.Error(fmt.Sprintf("log %d", i))
	}
}
