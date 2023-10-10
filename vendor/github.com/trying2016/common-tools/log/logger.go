package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Switch for log
var gLogTraceEnabled, gLogWarnEnabled, gLogInfoEnabled bool

// Async output to file
var fileLoggerQueue chan string
var jobWaiter sync.WaitGroup
var logFilePath string
var logFileSliceSize int64

// Log instance
var stdOut, stdErr *log.Logger
var LoggerOutput *LogWriter
// log writer
type LogWriter struct {
	out *os.File
}
func (l *LogWriter) Write(p []byte) (n int, err error){
	if fileLoggerQueue != nil {
		fileLoggerQueue <- string(p)
	}
	if l.out != nil {
		return l.out.Write(p)
	}
	return len(p), err
}

// Init logger
func Init(level, out string, size int64) error {
	if level == "warn" {
		gLogWarnEnabled = true
		gLogInfoEnabled = true
	} else if level == "info" {
		gLogInfoEnabled = true
	} else {
		gLogTraceEnabled = true
		gLogWarnEnabled = true
		gLogInfoEnabled = true
	}

	// logger writer
	LoggerOutput = &LogWriter{}

	outputs := strings.Split(strings.Trim(out, " "), ",")
	if len(outputs) > 0 {
		for i := 0; i < len(outputs); i++ {
			o := strings.Trim(outputs[i], " ")
			if o == "Console" && stdOut == nil {
				stdOut = log.New(os.Stdout, "", log.LstdFlags)
				stdErr = log.New(os.Stderr, "", log.LstdFlags)
				LoggerOutput.out = os.Stdout
			} else if len(o) > 0 && len(logFilePath) == 0 {
				logFilePath, _ = filepath.Abs(o)
			} else if len(o) > 0 {
				return fmt.Errorf(`Unacceptable output "%s"`, o)
			}
		}
	} else {
		// Default output to console
		stdOut = log.New(os.Stdout, "", log.LstdFlags)
		stdErr = log.New(os.Stderr, "", log.LstdFlags)
	}

	if len(logFilePath) > 0 {
		fileLoggerQueue = make(chan string, 1024)
		logFileSliceSize = size
		jobWaiter.Add(1)
		go saveLogTask()
	}
	return nil
}

// UnInit for destory log
func UnInit() {
	if fileLoggerQueue != nil {
		fileLoggerQueue <- ""
	}

	jobWaiter.Wait()

	if fileLoggerQueue != nil {
		close(fileLoggerQueue)
		fileLoggerQueue = nil
	}
}

// Trace log
func Trace(format string, v ...interface{}) {
	if gLogTraceEnabled {
		s := "[T] " + fmt.Sprintf(format, v...) + "\r\n"
		if stdOut != nil {
			stdOut.Output(2, s)
		}
		if fileLoggerQueue != nil {
			fileLoggerQueue <- s
		}
	}
}

// Warn log
func Warn(format string, v ...interface{}) {
	if gLogWarnEnabled {
		s := "[W] " + fmt.Sprintf(format, v...) + "\r\n"
		if stdErr != nil {
			stdErr.Output(2, s)
		}
		if fileLoggerQueue != nil {
			fileLoggerQueue <- s
		}
	}
}

// Info log
func Info(format string, v ...interface{}) {
	if gLogInfoEnabled {
		s := "[I] " + fmt.Sprintf(format, v...) + "\r\n"
		if stdOut != nil {
			stdOut.Output(2, s)
		}
		if fileLoggerQueue != nil {
			fileLoggerQueue <- s
		}
	}
}

// Error log
func Error(format string, v ...interface{}) {
	s := "[E] " + fmt.Sprintf(format, v...) + "\r\n"
	if stdErr != nil {
		stdErr.Output(2, s)
	}
	if fileLoggerQueue != nil {
		fileLoggerQueue <- s
	}
}

func saveLogTask() {
	defer jobWaiter.Done()

	var loggerFile *os.File
	defer func() {
		if loggerFile != nil {
			loggerFile.Close()
		}
	}()

	var logger *log.Logger
	createFile := func() bool {
		f, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			if stdErr != nil {
				print(fmt.Sprintf("Create log file error, %v\n", err))
			}
			return false
		} else {
			loggerFile = f
			if fi, err := loggerFile.Stat(); err == nil && fi.Size() == 0 {
				loggerFile.WriteString("\r\n\r\n\r\n")
			}
			logger = log.New(loggerFile, "", log.LstdFlags)
		}
		return true
	}
	fileLoggerQueue <- " " // init env

	var sliceTick int
	for {
		select {
		case line := <-fileLoggerQueue:
			if line == "" {
				return
			}
			if loggerFile == nil {
				if !createFile() {
					continue
				}
			}
			if logger != nil {
				logger.Output(2, line)

				// Slice log file
				sliceTick++
				if sliceTick%200 == 0 {
					sliceTick = 0
					if fi, err := loggerFile.Stat(); err == nil && fi.Size() >= logFileSliceSize {
						loggerFile.Close()
						loggerFile = nil
						logger = nil

						// Rename to
						now := time.Now().Format("2006-01-02-15-04-05")
						if err := os.Rename(logFilePath, logFilePath+"."+now); err != nil {
							print(fmt.Sprintf("Rename log file error, %v\n", err))
						}

						createFile()
					}
				}
			}
		}
	}
}
