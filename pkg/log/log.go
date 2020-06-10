package log

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

func ErrorLogf(format string, v ...interface{}) {
	errorLog.Printf(format, v...)
}

func ErrorLog(v ...interface{}) {
	errorLog.Print(v...)
}

func ErrorLogLn(v ...interface{}) {
	errorLog.Println(v...)
}

func WarnLog(v ...interface{}) {
	warnLog.Print(v...)
}

func DebugLog(v ...interface{}) {
	debugLog.Print(v...)
}

func Print(v ...interface{}) {
	infoLog.Print(v...)
}

func Printf(format string, v ...interface{}) {
	infoLog.Printf(format, v...)
}

func Println(v ...interface{}) {
	infoLog.Println(v...)
}

var (
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
)

func InitLog(
	debugWriter io.Writer,
	infoWriter io.Writer,
	warningWriter io.Writer,
	errorWriter io.Writer) {

	if debugWriter != nil {
		debugLog = log.New(debugWriter, "Debug: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	if infoWriter != nil {
		infoLog = log.New(infoWriter, "Info: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	if warningWriter != nil {
		warnLog = log.New(warningWriter, "Warning: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	if errorWriter != nil {
		errorLog = log.New(errorWriter, "Error: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

// Log to current file
func DefaultLogsFile(logFilename string) error {
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	InitLog(logFile, logFile, logFile, logFile)

	return nil
}

func LogsToStdErr() {
	InitLog(os.Stderr, os.Stderr, os.Stderr, os.Stderr)
}

func DiscardLog() {
	InitLog(ioutil.Discard, ioutil.Discard, ioutil.Discard, ioutil.Discard)
}
