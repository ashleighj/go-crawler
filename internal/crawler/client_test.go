package crawler

import (
	"net/http"
	"strings"
	"testing"
	"webcrawler/test/util"
)

func TestFetchPageBody(t *testing.T) {
	tests := []struct {
		name                  string
		path                  string
		statusCode            int
		headers               map[string]string
		response              string
		expectedError         bool
		expectedErrorContains string
	}{
		{
			name:       "success",
			path:       "/test",
			statusCode: http.StatusOK,
			headers:    map[string]string{"Content-Type": "text/html; charset=UTF-8"},
			response: `
		<!doctype html>
		<html>
			<head>
				<meta charset="utf-8">
				<title>Test Webpage</title>
				<meta name="description" content="Test Webpage">
			</head>
			<body> 
				<a href="https://www.google.com">Test Link</a>
			</body>
		</html>`,
		},
		{
			name:                  "fail_get-error",
			path:                  "/",
			expectedError:         true,
			expectedErrorContains: "error fetching page",
		},
		{
			name:                  "fail_status-code",
			path:                  "/test",
			statusCode:            http.StatusForbidden,
			expectedError:         true,
			expectedErrorContains: "status code",
		},
		{
			name:                  "fail_content-type",
			path:                  "/api",
			statusCode:            http.StatusOK,
			response:              `{"data":"hello"}`,
			headers:               map[string]string{"Content-Type": "application/json"},
			expectedError:         true,
			expectedErrorContains: "no html in page",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			server := util.GetTestServer(test.path, test.statusCode, test.response, test.headers)
			defer server.Close()

			body, e := FetchPageBody(server.URL)

			if test.expectedError {
				if e == nil {
					t.Errorf("missing error, should contain - %s", test.expectedErrorContains)
				}
				if !strings.Contains(e.Error(), test.expectedErrorContains) {
					t.Errorf("error mismatch.\n- received: %s\n- expected to contain: %s", e, test.expectedErrorContains)
				}
			}
			if test.expectedError == false && e != nil {
				t.Errorf("unexpected error - %s", e)
			}
			if test.expectedError == false && body == nil {
				t.Error("response body should not be nil")
			}
		})
	}
}
