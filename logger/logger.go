package webcrawler

import (
	"fmt"
	"log"
)

const (
	messageFmt  = "type=\"%s\" message=\"%v\""
	prefixInfo  = "INFO"
	prefixWarn  = "WARN"
	prefixError = "ERROR"
	prefixFatal = "FATAL"
)

// Info - posts an info-level log
func Info(message string) {
	log.Printf(messageFmt, prefixInfo, message)
}

// Infof - posts a formatted info-level log
func Infof(formatString string, vars ...interface{}) {
	log.Printf(messageFmt, prefixInfo, fmt.Sprintf(formatString, vars...))
}

// Error - posts an error log
func Error(message interface{}) {
	log.Printf(messageFmt, prefixError, message)
}

// Errorf - posts a formatted error log
func Errorf(formatString string, vars ...interface{}) {
	log.Printf(messageFmt, prefixError, fmt.Sprintf(formatString, vars...))
}

// Warn - posts a warning log
func Warn(message interface{}) {
	log.Printf(messageFmt, prefixWarn, message)
}

// Warnf - posts a formatted warn-level log
func Warnf(formatString string, vars ...interface{}) {
	log.Printf(messageFmt, prefixWarn, fmt.Sprintf(formatString, vars...))
}

// Fatal - posts a fatal log and shuts down the application
func Fatal(message interface{}) {
	log.Fatalf(messageFmt, prefixFatal, message)
}
