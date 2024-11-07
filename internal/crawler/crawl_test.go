package crawler

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
	config "webcrawler/config/crawler"
	"webcrawler/internal/util"
	testutil "webcrawler/test/util"
)

func TestNewCrawlSession(t *testing.T) {
	tests := []struct {
		name           string
		expectedResult CrawlSession
	}{
		{
			name: "success",
			expectedResult: CrawlSession{
				ToBeFiltered: make(chan *Page),
				ToBeVisited:  make(chan *Page),
				HostChannels: make(map[string]chan *Page),
				VisitedURLs:  NewConcurrentMap(),
				SeenContent:  NewConcurrentMap(),
				PendingURLs:  NewConcurrentCounter(),
				DoneChan:     make(chan bool),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			session := NewCrawlSession(3)

			if session.ToBeFiltered == nil || len(session.ToBeFiltered) != len(test.expectedResult.ToBeFiltered) {
				t.Errorf("unexpected result.\n- received: %v\n- expected %v", session.ToBeFiltered, test.expectedResult.ToBeFiltered)
			}
			if session.ToBeVisited == nil || len(session.ToBeVisited) != len(test.expectedResult.ToBeVisited) {
				t.Errorf("unexpected result.\n- received: %v\n- expected %v", session.ToBeVisited, test.expectedResult.ToBeVisited)
			}
			if session.HostChannels == nil || len(session.HostChannels) != len(test.expectedResult.HostChannels) {
				t.Errorf("unexpected result.\n- received: %v\n- expected %v", session.HostChannels, test.expectedResult.HostChannels)
			}
			if session.VisitedURLs == nil {
				t.Errorf("unexpected result.\n- received: %v\n- expected %v", session.VisitedURLs, test.expectedResult.VisitedURLs)
			}
			if session.SeenContent == nil {
				t.Errorf("unexpected result.\n- received: %v\n- expected %v", session.SeenContent, test.expectedResult.SeenContent)
			}
			if session.PendingURLs == nil || session.PendingURLs.GetCount() != test.expectedResult.PendingURLs.GetCount() {
				t.Errorf("unexpected result.\n- received: %v\n- expected %v", session.PendingURLs, test.expectedResult.PendingURLs)
			}
			if session.DoneChan == nil || len(session.DoneChan) != len(test.expectedResult.DoneChan) {
				t.Errorf("unexpected result.\n- received: %v\n- expected %v", session.DoneChan, test.expectedResult.DoneChan)
			}
		})
	}
}

func TestFilterURLs(t *testing.T) {
	tests := []struct {
		name                 string
		pages                []*Page
		expectedPendingCount int
	}{
		{
			name: "success_all_accepted",
			pages: []*Page{{
				URL:   "http://www.google.com",
				Depth: 0,
			}, {
				URL:   "http://www.google.com/images",
				Depth: 1,
			}},
			expectedPendingCount: 2,
		},
		{
			name: "success_one_accepted_one_rejected",
			pages: []*Page{{
				URL:   "http://www.google.com",
				Depth: 0,
			}, {
				URL:   "http://www.google.com/images",
				Depth: config.Get().MaxDepth + 1,
			}},
			expectedPendingCount: 1,
		},
		{
			name: "success_none_accepted",
			pages: []*Page{{
				URL:   "http://www.google.com",
				Depth: config.Get().MaxDepth + 1,
			}, {
				URL:   "http://www.google.com/images",
				Depth: config.Get().MaxDepth + 1,
			}},
			expectedPendingCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			logBuffer := testutil.GetLogBuffer()

			session := NewCrawlSession(3)
			go session.FilterURLs()
			go func() {
				for {
					<-session.ToBeVisited
				}
			}()
			go func() {
				for {
					<-session.DoneChan
				}
			}()

			for _, page := range test.pages {
				session.ToBeFiltered <- page
			}

			time.Sleep(200 * time.Millisecond)
			t.Log(logBuffer.String())

			if session.PendingURLs.GetCount() != test.expectedPendingCount {
				t.Errorf("expectedPendingCount mismatch.\n- received: %d\n- expected %d", session.PendingURLs.GetCount(), test.expectedPendingCount)
			}
		})
	}
}

func TestRouteAcceptedURLs(t *testing.T) {
	tests := []struct {
		name  string
		pages []*Page
	}{
		{
			name: "success_all_accepted",
			pages: []*Page{{
				URL: "http://www.google.com",
			}, {
				URL: "http://www.google.com/images",
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			logBuffer := testutil.GetLogBuffer()

			session := NewCrawlSession(3)
			go session.RouteAcceptedURLs()

			for _, page := range test.pages {
				go func(page *Page) {
					for {
						channel, _ := session.GetHostChannel(page)
						<-channel
					}
				}(page)
			}

			for _, page := range test.pages {
				session.ToBeVisited <- page
			}

			time.Sleep(200 * time.Millisecond)
			t.Log(logBuffer.String())

			for _, page := range test.pages {
				domain, e := GetURLDomain(page.URL)
				if e != nil {
					t.Errorf("bad test input, page url - %s", page.URL)
				}

				if _, ok := session.HostChannels[domain]; !ok {
					t.Errorf("no host-specific channel found for domain [%s], input url [%s]", domain, page.URL)
				}
			}
		})
	}
}

func TestCheckDone(t *testing.T) {
	session := NewCrawlSession(3)

	isDone := false

	session.PendingURLs.Add(1)
	session.CheckDone()
	time.Sleep(200 * time.Millisecond)

	if isDone {
		t.Error("check done not working - 'isDone' should be false")
	}

	go func(done *bool) {
		for {
			<-session.DoneChan
			*done = true
		}
	}(&isDone)

	session.PendingURLs.Subtract(1)
	session.CheckDone()
	time.Sleep(200 * time.Millisecond)

	if !isDone {
		t.Error("check done not working - 'isDone' should be true")
	}
}

func TestGetHostChannel(t *testing.T) {

	tests := []struct {
		name          string
		page          *Page
		errorExpected bool
	}{
		{
			name: "success_new_chan",
			page: &Page{
				URL: "https://www.google.com",
			},
		},
		{
			name: "success_existing_chan",
			page: &Page{
				URL: "https://www.google.com",
			},
		},
		{
			name: "fail_bad_url",
			page: &Page{
				URL: "https://www.google.com/",
			},
			errorExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			logBuffer := testutil.GetLogBuffer()

			session := NewCrawlSession(3)
			hostChan, e := session.GetHostChannel(test.page)

			t.Log(logBuffer.String())

			if test.errorExpected && e == nil {
				t.Errorf("missing expected error")
			}

			if !test.errorExpected {
				if e != nil {
					t.Errorf("unexpected error - %s", e)
				}
				if hostChan == nil {
					t.Error("missing expected return value")
				}
				domain, _ := GetURLDomain(test.page.URL)
				if _, ok := session.HostChannels[domain]; !ok {
					t.Error("expected host channel not in host channel map")
				}
			}
		})
	}
}

func TestCrawl(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		pageContent      string
		isSeen           bool
		expectedVisited  []string
		expectedChildren []string
	}{
		{
			name: "success",
			path: "/",
			pageContent: `
		<!doctype html>
		<html>
			<head>
				<meta charset="utf-8">
				<title>Test Webpage</title>
				<meta name="description" content="Test Webpage">
			</head>
			<body> 
				<a href="/test">Test Link</a>
			</body>
		</html>`,
			expectedVisited:  []string{"/"},
			expectedChildren: []string{"/test"},
		},
		{
			name:             "fail_broken_link",
			path:             "/",
			expectedVisited:  []string{},
			expectedChildren: []string{},
		},
		{
			name:             "fail_empty_pageBody",
			path:             "/",
			pageContent:      "",
			expectedVisited:  []string{"/"},
			expectedChildren: []string{},
		},
		{
			name:   "success_already_seen",
			path:   "/",
			isSeen: true,
			pageContent: `
		<!doctype html>
		<html>
			<head>
				<meta charset="utf-8">
				<title>Test Webpage</title>
				<meta name="description" content="Test Webpage">
			</head>
			<body> 
				<a href="/test">Test Link</a>
			</body>
		</html>`,
			expectedVisited:  []string{},
			expectedChildren: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			logBuffer := testutil.GetLogBuffer()

			server := testutil.GetTestServer(test.path, http.StatusOK, test.pageContent, map[string]string{"Content-Type": "text/html"})
			defer server.Close()

			session := NewCrawlSession(3)
			session.PendingURLs.Add(1)

			domain, _ := GetURLDomain(server.URL)
			session.HostChannels[domain] = make(chan *Page)

			if test.isSeen {
				session.SeenContent.Add(util.Hash(test.pageContent), 1)
			}

			go session.CrawlDomainURLs(domain, session.HostChannels[domain])

			go func() {
				for {
					<-session.ToBeFiltered
				}
			}()

			url := fmt.Sprintf("%s%s", server.URL, test.path)
			page := NewPage(url, url, 0, nil)

			go func() {
				session.HostChannels[domain] <- page
			}()

			<-session.DoneChan

			t.Log(logBuffer.String())

			for _, link := range test.expectedVisited {
				if !session.VisitedURLs.KeyExists(page.URLHash) {
					t.Errorf("missing expected visited link [%s]", link)
				}
			}

			if len(session.VisitedURLs.data) != len(test.expectedVisited) {
				t.Errorf("visited links mismatch.\n- received: %d\n- expected: %d", len(session.VisitedURLs.data), len(test.expectedVisited))
			}

			for _, expected := range test.expectedChildren {
				found := false
				for _, child := range page.Children {
					if child.URL == fmt.Sprintf("%s%s", server.URL, expected) {
						found = true
					}
				}
				if !found {
					t.Errorf("missing expected page child [%s]", expected)
				}
			}
			
			if len(test.expectedChildren) != len(page.Children) {
				t.Errorf("page children mismatch.\n- received: %d\n- expected: %d", len(page.Children), len(test.expectedChildren))
			}
		})
	}
}

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

			server := testutil.GetTestServer(test.path, test.statusCode, test.response, test.headers)
			defer server.Close()

			session := NewCrawlSession(3)
			body, e := session.FetchPageBody(server.URL)

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
