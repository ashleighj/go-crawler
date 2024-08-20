package main

import (
	"fmt"

	crawlerConfig "webcrawler/config/crawler"
	crawler "webcrawler/internal/crawler"
	logger "webcrawler/logger"
)

/*
   Dear Reviewer, your time and effort looking over my code is super appreciated.
	 I know you probably have a lot of other things going on!
	 To hopefully save you some of that time & effort, I have added a ton of unusually-granular comments
	 in an attempt to help you get through understanding my intentions/logic faster.
	 Hope that helps!
*/

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

	crawlerSession := crawler.NewCrawlSession()

	// fetch seed urls from config
	crawlerConfig := crawlerConfig.Get()

	// check seed urls are not empty
	if len(crawlerConfig.Seeds) == 0 {
		logger.Error("no configured seeds, nowhere to crawl :(")
		crawlerSession.DoneChan <- true
	}

	// decide which urls are appropriate to crawl
	go crawlerSession.FilterURLs()

	// route filtered urls to host-specific channel
	go crawlerSession.RouteAcceptedUrls()

	// send seed urls to be filtered and crawled
	for _, url := range crawlerConfig.Seeds {
		page := crawler.NewPage(url, url, 0, nil)
		defer page.PrintTree()
		crawlerSession.ToBeFiltered <- page
	}

	<-crawlerSession.DoneChan
}
