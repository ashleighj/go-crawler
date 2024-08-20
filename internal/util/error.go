package util

import (
	logger "webcrawler/logger"
)

func CheckErr(e error) {
	if e != nil {
		logger.Error(e)
	}
}

func CheckErrFatal(e error) {
	if e != nil {
		logger.Fatal(e)
	}
}
