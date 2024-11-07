package main

import (
	"fmt"

	crawlerConfig "webcrawler/config/crawler"
	crawler "webcrawler/internal/crawler"
	logger "webcrawler/logger"
)

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

	// fetch seed urls from config
	crawlerConfig := crawlerConfig.Get()

	crawlerSession := crawler.NewCrawlSession(crawlerConfig.ReadTimeoutSeconds)

	// check seed urls are not empty
	if len(crawlerConfig.Seeds) == 0 {
		logger.Error("no configured seeds, nowhere to crawl :(")
		crawlerSession.DoneChan <- true
	}

	// decide which urls are appropriate to crawl
	go crawlerSession.FilterURLs()

	// route filtered urls to host-specific channel
	go crawlerSession.RouteAcceptedURLs()

	// send seed urls to be filtered and crawled
	for _, url := range crawlerConfig.Seeds {
		page := crawler.NewPage(url, url, 0, nil)
		defer page.PrintTree()
		crawlerSession.ToBeFiltered <- page
	}

	<-crawlerSession.DoneChan
}
