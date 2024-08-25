package util

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	logger "webcrawler/logger"
)

func GetLogBuffer() *bytes.Buffer {
	var buffer bytes.Buffer
	log.SetOutput(&buffer)
	return &buffer
}

func GetTestServer(
	path string,
	statusCode int,
	responseBody string,
	headers map[string]string) *httptest.Server {

	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			logger.Infof("test server hit - path [%s]", path)

			r.URL.Path = path
			for key, val := range headers {
				w.Header().Add(key, val)
			}

			w.WriteHeader(statusCode)

			if responseBody != "" {
				w.Write([]byte(responseBody))
			}
		}))
}
