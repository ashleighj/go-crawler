package webcrawler

import (
	"fmt"
	"log"
	"strings"
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
	log.Printf(messageFmt, prefixInfo, escape(message))
}

// Infof - posts a formatted info-level log
func Infof(formatString string, vars ...interface{}) {
	log.Printf(messageFmt, prefixInfo, escape(fmt.Sprintf(formatString, vars...)))
}

// Error - posts an error log
func Error(message interface{}) {
	log.Printf(messageFmt, prefixError, message)
}

// Errorf - posts a formatted error log
func Errorf(formatString string, vars ...interface{}) {
	log.Printf(messageFmt, prefixError, escape(fmt.Sprintf(formatString, vars...)))
}

// Warn - posts a warning log
func Warn(message interface{}) {
	log.Printf(messageFmt, prefixWarn, message)
}

// Warnf - posts a formatted warn-level log
func Warnf(formatString string, vars ...interface{}) {
	log.Printf(messageFmt, prefixWarn, escape(fmt.Sprintf(formatString, vars...)))
}

// Fatal - posts a fatal log and shuts down the application
func Fatal(message interface{}) {
	escaped := escape(fmt.Sprintf("%v", message))
	log.Fatalf(messageFmt, prefixFatal, escaped)
}

func escape(input string) string {
	input = strings.ReplaceAll(input, "\"", "\\\"")
	return strings.ReplaceAll(input, "\\", "\\\\")
}
