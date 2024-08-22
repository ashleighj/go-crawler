package util

import (
	"bytes"
	"log"
)

func GetLogBuffer() *bytes.Buffer {
	var buffer bytes.Buffer
	log.SetOutput(&buffer)
	return &buffer
}
