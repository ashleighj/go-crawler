package main

import (
	"bytes"
	"fmt"
	"io"
	"time"

	crawlerConfig "webcrawler/config/crawler"
	crawler "webcrawler/internal/crawler"
	"webcrawler/internal/util"
	logger "webcrawler/logger"
)

/*
   Dear Reviewer, your time and effort looking over my code is super appreciated.
	 I know you probably have a lot of other things going on!
	 To hopefully save you some of that time & effort, I have added a ton of unusually-granular comments
	 in an attempt to help you get through understanding my intentions/logic faster.
	 Hope that helps!
*/

func filter(
	toBeFiltered chan *crawler.Page,
	toBeVisited chan *crawler.Page,
	visitedURLs *crawler.ConcurrentMap,
	seenContent *crawler.ConcurrentMap,
	pendingURLs *crawler.ConcurrentCounter,
	doneChan chan bool) {

	for {
		select {
		case page := <-toBeFiltered:
			logger.Infof("new page to be filtered - %s", page.URL)

			if page.IsCrawlable(visitedURLs, seenContent) {
				logger.Infof("new page accepted - %s", page.URL)
				pendingURLs.Add(1)
				toBeVisited <- page
			} else {
				logger.Infof("new page rejected - %s", page.URL)
				checkDone(pendingURLs, doneChan)
			}
		default:
		}
	}
}

// route routes the url to a channel based on its host
func route(
	toBeVisited chan *crawler.Page,
	hostChanMap map[string]chan *crawler.Page,
	pendingURLs *crawler.ConcurrentCounter,
	visitedURLs *crawler.ConcurrentMap,
	seenContent *crawler.ConcurrentMap,
	toBeFiltered chan *crawler.Page,
	doneChan chan bool) {

	for {
		select {
		case page := <-toBeVisited:
			logger.Infof("new page to be routed - %s", page.URL)
			channel := getHostChannel(page, hostChanMap, pendingURLs, visitedURLs, seenContent, toBeFiltered, doneChan)
			channel <- page
		default:
		}
	}
}

func getHostChannel(
	page *crawler.Page,
	hostChanMap map[string]chan *crawler.Page,
	pendingURLs *crawler.ConcurrentCounter,
	visitedURLs *crawler.ConcurrentMap,
	seenContent *crawler.ConcurrentMap,
	toBeFiltered chan *crawler.Page,
	doneChan chan bool) (hostChan chan *crawler.Page) {

	// get url domain part
	domain, e := crawler.GetURLDomain(page.URL)
	if e != nil {
		logger.Warnf("could not get url domain in order to find host channel - %s", e)
		return
	}

	// check if domain exists in map & if not, create channel entry
	channel, ok := hostChanMap[domain]
	if !ok {
		channel = make(chan *crawler.Page)
		hostChanMap[domain] = channel

		go crawlDomainURLs(domain, channel, pendingURLs, visitedURLs, seenContent, toBeFiltered, doneChan)
		logger.Infof("host channel created for domain [%s]", domain)
	}

	logger.Infof("returning host channel for domain [%s]", domain)
	return channel
}

func crawlDomainURLs(
	domain string,
	channel chan *crawler.Page,
	pendingURLs *crawler.ConcurrentCounter,
	visitedURLs *crawler.ConcurrentMap,
	seenContent *crawler.ConcurrentMap,
	toBeFiltered chan *crawler.Page,
	doneChan chan bool) {

	logger.Infof("now receiving urls to be crawled from domain [%s]", domain)

	for {
		page := <-channel
		time.Sleep(time.Duration(crawlerConfig.Get().DomainHitDelayMS) * time.Millisecond)

		logger.Infof("queueing new link [%s] from domain [%s] for crawl", page.URL, domain)

		go crawl(page, pendingURLs, visitedURLs, seenContent, toBeFiltered, doneChan)
	}
}

func crawl(
	currentPage *crawler.Page,
	pendingURLs *crawler.ConcurrentCounter,
	visitedURLs *crawler.ConcurrentMap,
	seenContent *crawler.ConcurrentMap,
	toBeFiltered chan *crawler.Page,
	doneChan chan bool) {

	var children []*crawler.Page
	defer func() {
		pendingURLs.Subtract(1)
		checkDone(pendingURLs, doneChan)
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
	pageBody, e := crawler.FetchPageBody(currentPage.URL)
	if e != nil {
		logger.Warnf("broken link [%s], can't crawl - %s", currentPage.URL, e)
		return
	}
	if pageBody == nil {
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

	if seenContent.KeyExists(currentPage.ContentHash) {
		logger.Infof("page [%s] not crawlable - content already seen", currentPage.URL)
		return
	}

	visitedURLs.Add(currentPage.URLHash, 1)
	seenContent.Add(currentPage.ContentHash, 1)

	children = currentPage.GetChildren(pageBody, currentPage.Depth+1)
	currentPage.Children = children

	for _, child := range children {
		toBeFiltered <- child
	}
}

func checkDone(pendingURLs *crawler.ConcurrentCounter, doneChan chan bool) {
	logger.Info("checking if done")

	if pendingURLs.GetCount() == 0 {
		logger.Info("no more pending urls, ending crawl")
		doneChan <- true
		return
	}

	logger.Infof("pending url count [%d]", pendingURLs.GetCount())
	logger.Info("crawl continuing...")
}

func main() {
	fmt.Print(
		`
     .'(   )\.---.     /(,-.          )\.-.     /'-.     /'-.        .'(   .')       )\.---.     /'-.  
 ,') \  ) (   ,-._(  ,' _   )       ,' ,-,_)  ,' _  \  ,' _  \   ,') \  ) ( /       (   ,-._(  ,' _  \ 
(  /(/ /   \  '-,   (  '-' (       (  .   _  (  '-' ( (  '-' (  (  /(/ /   ))        \  '-,   (  '-' ( 
 )    (     ) ,-'    )  _   )       ) '..' )  ) ,_ .'  )   _  )  )    (    )'._.-.    ) ,-'    ) ,_ .' 
(  .'\ \   (  ''-.  (  '-' /       (  ,   (  (  ' ) \ (  ,' ) \ (  .'\ \  (       )  (  ''-.  (  ' ) \ 
 )/   )/    )..-.(   )/._.'         )/'._.'   )/   )/  )/    )/  )/   )/   )/,__.'    )..-.(   )/   )/ 
                                                                                                      
`)

	// waits for exit signal - keeps main go routine running until an appropriate shutdown time
	doneChan := make(chan bool)

	// all urls yet to be directed or downloaded
	toBeFiltered := make(chan *crawler.Page)

	// all urls yet to be directed or downloaded
	toBeVisited := make(chan *crawler.Page)

	// enforce politeness by having separate goroutines process urls per host,
	// so number of host visits within a specific timeframe can be controlled
	hostChanMap := make(map[string]chan *crawler.Page)

	// enables safe counting of urls still to be crawled
	pendingURLs := crawler.NewConcurrentCounter()

	// stores hashes of links already crawled
	visitedURLs := crawler.NewConcurrentMap()

	// stores hashes of the content of pages already crawled
	seenContent := crawler.NewConcurrentMap()

	// fetch seed urls from config
	crawlerConfig := crawlerConfig.Get()

	// check seed urls are not empty
	if len(crawlerConfig.Seeds) == 0 {
		logger.Error("no configured seeds, nowhere to crawl :(")
		doneChan <- true
	}

	// decide which urls are appropriate to crawl
	go filter(toBeFiltered, toBeVisited, visitedURLs, seenContent, pendingURLs, doneChan)

	// route filtered urls to host-specific channel
	go route(toBeVisited, hostChanMap, pendingURLs, visitedURLs, seenContent, toBeFiltered, doneChan)

	// send seed urls to be filtered and crawled
	for _, url := range crawlerConfig.Seeds {
		page := crawler.NewPage(url, url, 0, nil)
		defer page.PrintTree()
		toBeFiltered <- page
	}

	<-doneChan
}
