package logger

import (
	"fmt"
)

type PrintLogger struct {
	prefix 		string
	logLevel 	LogLevel
}

func NewPrintLogger(prefix string, logLevel LogLevel) *PrintLogger {
	return &PrintLogger{
		prefix: prefix,
		logLevel: logLevel,
	}
}

func (this *PrintLogger) Debug(str string) {
	if this.logLevel <= DEBUG {
		fmt.Printf("[DEBUG] %v: %v\n",  this.prefix, str)
	}
}

func (this *PrintLogger) Info(str string) {
	if this.logLevel <= INFO {
		fmt.Printf("[INFO] %v: %v\n",  this.prefix, str)
	}
}

func (this *PrintLogger) Warn(str string) {
	if this.logLevel <= WARNING {
		fmt.Printf("[WARN] %v: %v\n",  this.prefix, str)
	}
}

func (this *PrintLogger) Error(str string) {
	if this.logLevel <= ERROR {
		fmt.Printf("[ERROR] %v: %v\n",  this.prefix, str)
	}
}
