package log

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

const (
	Discard = 1 << iota
	Stdout
	Stderr
	EnableFile
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func Config(flagTrace int, flagInfo int, flagWarning int, flagError int, filePath string) {
	isEnableFile := false
	var file *os.File
	file, isEnableFile, Trace = config(flagTrace, filePath, file, isEnableFile, "TRACE: ")
	file, isEnableFile, Info = config(flagInfo, filePath, file, isEnableFile, "INFO: ")
	file, isEnableFile, Warning = config(flagWarning, filePath, file, isEnableFile, "WARNING: ")
	file, isEnableFile, Error = config(flagError, filePath, file, isEnableFile, "ERROR: ")
}

func config(flag int, filePath string, file *os.File, isEnableFile bool, prefix string) (*os.File, bool, *log.Logger) {
	var temp *log.Logger
	switch flag {
	case Discard:
		temp = log.New(ioutil.Discard, prefix, log.Ldate|log.Ltime|log.Lshortfile)
	case Stdout:
		temp = log.New(os.Stdout, prefix, log.Ldate|log.Ltime|log.Lshortfile)
	case Stderr:
		temp = log.New(os.Stderr, prefix, log.Ldate|log.Ltime|log.Lshortfile)
	case Stdout | EnableFile:
		file = enableFile(filePath)
		isEnableFile = true
		temp = log.New(io.MultiWriter(file, os.Stdout), prefix, log.Ldate|log.Ltime|log.Lshortfile)
	case Stderr | EnableFile:
		if !isEnableFile {
			file = enableFile(filePath)
			isEnableFile = true
		}
		temp = log.New(io.MultiWriter(file, os.Stderr), prefix, log.Ldate|log.Ltime|log.Lshortfile)
	default:
		log.Panicln("log,配置错误")

	}
	return file, isEnableFile, temp
}

func enableFile(filePath string) *os.File {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}
	return file
}
