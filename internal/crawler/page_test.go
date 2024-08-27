package crawler

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	config "webcrawler/config/crawler"
	"webcrawler/internal/util"
)

func TestIsCrawlable(t *testing.T) {
	tests := []struct {
		name           string
		page           Page
		blacklist      []string
		pageVisited    bool
		contentSeen    bool
		expectedResult bool
	}{
		{
			name: "success_true",
			page: Page{
				URL:   "https://www.google.com",
				Depth: 0,
			},
			expectedResult: true,
		},
		{
			name: "success_false_maxdepth",
			page: Page{
				URL:   "https://www.google.com",
				Depth: config.Get().MaxDepth,
			},
			expectedResult: false,
		},
		{
			name: "success_false_blacklisted",
			page: Page{
				URL:   "https://www.google.com",
				Depth: 0,
			},
			blacklist:      []string{"https://www.google.com"},
			expectedResult: false,
		},
		{
			name: "success_false_page_visited",
			page: Page{
				URL:     "https://www.google.com",
				Depth:   0,
				URLHash: util.Hash("www.google.com"),
			},
			pageVisited:    true,
			expectedResult: false,
		},
		{
			name: "success_false_content_seen",
			page: Page{
				URL:         "https://www.google.com",
				Depth:       0,
				ContentHash: util.Hash("<html><body>Hello World</body></html>"),
			},
			contentSeen:    true,
			expectedResult: false,
		},
		{
			name: "fail_url_form",
			page: Page{
				URL:   "https://www.google.com/",
				Depth: 0,
			},
			expectedResult: false,
		},
		{
			name: "fail_blaclisted_url_form",
			page: Page{
				URL:   "https://www.google.com",
				Depth: 0,
			},
			blacklist:      []string{"https://www.google.com/"},
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			session := NewCrawlSession()

			if test.pageVisited {
				session.VisitedURLs.Add(test.page.URLHash, 1)
			}
			if test.contentSeen {
				session.SeenContent.Add(test.page.ContentHash, 1)
			}
			if test.blacklist != nil && len(test.blacklist) > 0 {
				config.Get().BlacklistedURLs = append(config.Get().BlacklistedURLs, test.blacklist...)
			}

			result := test.page.IsCrawlable(session.VisitedURLs, session.SeenContent)

			if test.expectedResult != result {
				t.Errorf("result mismatch.\n- received: %t\n- expected: %t", result, test.expectedResult)
			}
		})
	}

}

func TestPrintTree(t *testing.T) {
	expected := `

[0] https://www.google.com                                      [1] https://images.google.com                                   
                                                                [1] https://news.google.com                                     `

	parent := NewPage("https://www.google.com", "google", 0, nil)
	children := []*Page{
		NewPage("https://images.google.com", "images", 1, parent),
		NewPage("https://news.google.com", "news", 1, parent)}
	parent.Children = children

	pipeReader, pipeWriter, _ := os.Pipe()
	stdOut := os.Stdout
	os.Stdout = pipeWriter

	parent.PrintTree()
	time.Sleep(100 * time.Millisecond)

	pipeWriter.Close()
	out, _ := io.ReadAll(pipeReader)
	os.Stdout = stdOut

	if fmt.Sprintf("%s", out) != expected {
		t.Errorf("output mismatch.\n- received:||%s||\n- expected:||%s||", out, expected)
	}
}
