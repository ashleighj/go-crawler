package crawler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	crawlerConfig "webcrawler/config/crawler"
	"webcrawler/internal/util"
	logger "webcrawler/logger"
)

// CrawlSession holds all data structures required to crawl a given set of seed URLs
type CrawlSession struct {

	// http client for crawling links
	Client http.Client

	// all urls yet to be directed or downloaded
	ToBeFiltered chan *Page

	// all urls yet to be directed or downloaded
	ToBeVisited chan *Page

	// enforces politeness by having separate goroutines process urls per host,
	// so number of host visits within a specific timeframe can be controlled
	HostChannels map[string]chan *Page

	// stores hashes of links already crawled
	VisitedURLs *ConcurrentMap

	// stores hashes of the content of pages already crawled
	SeenContent *ConcurrentMap

	// enables safe counting of urls still to be crawled
	PendingURLs *ConcurrentCounter

	// waits for exit signal - keeps main go routine running until an appropriate shutdown time
	DoneChan chan bool
}

// NewCrawlSession creates and returns a pointer to a new CrawlerSession struct
func NewCrawlSession(readTimeoutSecs int) *CrawlSession {
	return &CrawlSession{
		Client:       *&http.Client{Timeout: time.Duration(readTimeoutSecs) * time.Second},
		ToBeFiltered: make(chan *Page),
		ToBeVisited:  make(chan *Page),
		HostChannels: make(map[string]chan *Page),
		VisitedURLs:  NewConcurrentMap(),
		SeenContent:  NewConcurrentMap(),
		PendingURLs:  NewConcurrentCounter(),
		DoneChan:     make(chan bool)}
}

// FilterURLs continuously receives from the "ToBeFiltered" channel and decides
// which urls received should be send to the router for crawling
func (c *CrawlSession) FilterURLs() {
	for {
		select {
		case page := <-c.ToBeFiltered:
			logger.Infof("new page to be filtered - %s", page.URL)

			if page.IsCrawlable(c.VisitedURLs, c.SeenContent) {
				logger.Infof("new page accepted - %s", page.URL)

				c.PendingURLs.Add(1)
				c.ToBeVisited <- page
			} else {
				logger.Infof("new page rejected - %s", page.URL)
				c.CheckDone()
			}

		default:
		}
	}
}

// RouteAcceptedURLs received recently-filtered urls from the "ToBeVisited" channel and
// finds the appropriate channel to send them to for crawling based on their host. Crawling
// is split by host so that the timing of hits to that host can be controlled so as not to
// overwhelm it / break its rate-limiting rules
func (c *CrawlSession) RouteAcceptedURLs() {
	for {
		select {
		case page := <-c.ToBeVisited:
			logger.Infof("new page to be routed - %s", page.URL)

			channel, e := c.GetHostChannel(page)
			if e != nil {
				logger.Errorf("could not get host-specific channel to send page [%s] to", page.URL)
				break
			}

			channel <- page

		default:
		}
	}
}

// CheckDone checks whether it's time to finish the crawl session and print the link tree(s)
func (c *CrawlSession) CheckDone() {
	logger.Info("checking if done")

	if c.PendingURLs.GetCount() == 0 {
		logger.Info("no more pending urls, ending crawl")
		c.DoneChan <- true
		return
	}

	logger.Infof("pending url count [%d]", c.PendingURLs.GetCount())
	logger.Info("crawl continuing...")
}

// GetHostChannel finds the appropriate channel to send a crawlable page to in order to be crawled
func (c *CrawlSession) GetHostChannel(page *Page) (hostChannel chan *Page, e error) {
	// get url domain part
	domain, e := GetURLDomain(page.URL)
	if e != nil {
		e = fmt.Errorf("could not get url domain in order to find host channel - %s", e)
		return
	}

	// check if domain exists in map & if not, create channel entry
	channel, ok := c.HostChannels[domain]
	if !ok {
		channel = make(chan *Page)
		c.HostChannels[domain] = channel

		go c.CrawlDomainURLs(domain, channel)
		logger.Infof("host channel created for domain [%s]", domain)
	}

	logger.Infof("returning host channel for domain [%s]", domain)
	return channel, nil
}

// CrawlDomainURLs runs once per domain
func (c *CrawlSession) CrawlDomainURLs(domain string, channel chan *Page) {
	logger.Infof("now receiving urls to be crawled from domain [%s]", domain)

	for {
		page := <-channel

		logger.Infof("received new link [%s] from domain [%s] for crawl", page.URL, domain)
		time.Sleep(time.Duration(crawlerConfig.Get().DomainHitDelayMS) * time.Millisecond)
		logger.Infof("queueing new link [%s] from domain [%s] for crawl", page.URL, domain)

		go c.Crawl(page)
	}
}

// Crawl fetches content for the current Page and sends any new, valid links to be filtered for crawling
// In order for a link to be valid for "queuing" at this stage, it must not have previously been crawled, 
// and it must not be part of a page of content that has previously been seen perhaps under a different link
func (c *CrawlSession) Crawl(currentPage *Page) {
	var children []*Page

	defer func() {
		c.PendingURLs.Subtract(1)
		c.CheckDone()
	}()

	if currentPage == nil {
		logger.Error("current page nil, not crawlable")
		return
	}

	if currentPage.Parent != nil {
		logger.Infof("crawling page [%s], child of [%s], depth [%d]",
			currentPage.URL, currentPage.Parent.URL, currentPage.Depth)
	} else {
		logger.Infof("crawling seed page [%s], depth [%d]", currentPage.URL, currentPage.Depth)
	}

	// fetch page body
	pageBody, e := c.FetchPageBody(currentPage.URL)
	if e != nil {
		logger.Warnf("broken link [%s], can't crawl - %s", currentPage.URL, e)
		return
	}
	if pageBody == nil || pageBody == http.NoBody {
		logger.Errorf("page body for url [%s] is empty, nothing to crawl", currentPage.URL)
		return
	}

	// can't read the response twice, so after we've extracted the content to get the page content string,
	// we reload the response with those same bytes in order to create the page tokeniser below
	pageBytes, _ := io.ReadAll(pageBody)
	pageBody = io.NopCloser(bytes.NewBuffer(pageBytes))
	pageContentString := string(pageBytes)

	currentPage.RawContent = pageContentString
	currentPage.ContentHash = util.Hash(pageContentString)

	if c.SeenContent.KeyExists(currentPage.ContentHash) {
		logger.Infof("page [%s] not crawlable - content already seen", currentPage.URL)
		return
	}

	c.VisitedURLs.Add(currentPage.URLHash, 1)
	c.SeenContent.Add(currentPage.ContentHash, 1)

	children = currentPage.GetChildren(pageBody, currentPage.Depth+1)
	currentPage.Children = children

	for _, child := range children {
		c.ToBeFiltered <- child
	}
}

// FetchPageBody performs a GET request on the given url and returns
// the response body when it is of type "text/html"
func (c *CrawlSession) FetchPageBody(url string) (body io.ReadCloser, e error) {
	logger.Infof("fetching page [%s]", url)

	req, e := http.NewRequest(http.MethodGet, url, nil)
	if e != nil {
		e = fmt.Errorf("error creating GET request for url [%s] - %s", url, e)
		logger.Error(e)
		return
	}

	response, e := c.Client.Do(req)
	if e != nil {
		e = fmt.Errorf("error fetching page [%s] - %s", url, e)
		logger.Error(e)
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
