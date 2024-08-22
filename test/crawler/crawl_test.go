package crawler

import (
	"testing"
	"time"
	config "webcrawler/config/crawler"
	"webcrawler/internal/crawler"
	"webcrawler/test/util"
)

func TestNewCrawlSession(t *testing.T) {
	tests := []struct {
		name           string
		expectedResult crawler.CrawlSession
	}{
		{
			name: "success_standard",
			expectedResult: crawler.CrawlSession{
				ToBeFiltered: make(chan *crawler.Page),
				ToBeVisited:  make(chan *crawler.Page),
				HostChannels: make(map[string]chan *crawler.Page),
				VisitedURLs:  crawler.NewConcurrentMap(),
				SeenContent:  crawler.NewConcurrentMap(),
				PendingURLs:  crawler.NewConcurrentCounter(),
				DoneChan:     make(chan bool),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			session := crawler.NewCrawlSession()

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
		pages                []*crawler.Page
		expectedPendingCount int
	}{
		{
			name: "success_all_accepted",
			pages: []*crawler.Page{{
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
			pages: []*crawler.Page{{
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
			pages: []*crawler.Page{{
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

			logBuffer := util.GetLogBuffer()

			session := crawler.NewCrawlSession()
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
		pages []*crawler.Page
	}{
		{
			name: "success_all_accepted",
			pages: []*crawler.Page{{
				URL: "http://www.google.com",
			}, {
				URL: "http://www.google.com/images",
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			logBuffer := util.GetLogBuffer()

			session := crawler.NewCrawlSession()
			go session.RouteAcceptedURLs()

			for _, page := range test.pages {
				go func(page *crawler.Page) {
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
				domain, e := crawler.GetURLDomain(page.URL)
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

	session := crawler.NewCrawlSession()

	session.PendingURLs.Add(1)

	go func() {
		for {
			<-session.DoneChan
		}
	}()
}

func TestGetHostChannel(t *testing.T) {

}

func TestCrawlDomainURLs(t *testing.T) {

}

func TestCrawl(t *testing.T) {

}
