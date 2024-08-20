package crawler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	logger "webcrawler/logger"
)

func FetchPageBody(url string) (body io.ReadCloser, e error) {
	logger.Infof("fetching page [%s]", url)
	response, e := http.Get(url)
	if e != nil {
		logger.Errorf("error fetching page [%s - %s]", url, e)
		return
	}

	status := response.StatusCode
	if status < 200 || status > 299 {
		e = fmt.Errorf("could not fetch page [%s], status code [%d]", url, status)
		logger.Error(e)
		return
	}

	for _, contentType := range response.Header["Content-Type"] {
		if strings.Contains(contentType, "text/html") {
			return response.Body, nil
		}
	}
	e = fmt.Errorf("no html in page [%s]", url)
	logger.Error(e)
	return
}
