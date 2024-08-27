package crawler

import (
	"fmt"
	"io"
	"strings"
	"webcrawler/internal/util"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	crawlerConfig "webcrawler/config/crawler"
	logger "webcrawler/logger"
)

// Page holds data pertaining to a page in the site map of a configured seed url
type Page struct {
	URL         string
	LinkText    string
	URLHash     string
	RawContent  string
	ContentHash string
	Parent      *Page
	Children    []*Page
	Depth       int
}

// NewPage creates and returns a new page struct
func NewPage(url string, linkText string, depth int, parent *Page) *Page {
	url = TrimLinkVars(url)
	newPage := Page{
		URL:      url,
		LinkText: linkText,
		URLHash:  util.Hash(url),
		Depth:    depth,
		Parent:   parent}

	return &newPage
}

// GetChildren finds the links in a page, uses them to contruct new
// Page structs relating to the parent, and returns those new Page structs
func (page *Page) GetChildren(pageBody io.ReadCloser, depth int) (children []*Page) {
	logger.Infof("parsing page at [%s], finding children links", page.URL)
	// split page into tokens
	tokeniser := html.NewTokenizer(pageBody)

	var linkTagStart *html.Token
	linkTagText := ""

	// scan page content and collect subpages
	// loop until we find and error, which could also represent the end of the page stream
	for {
		tokenType := tokeniser.Next()

		if tokenType == html.ErrorToken {
			e := tokeniser.Err()
			if e != io.EOF {
				logger.Warnf("error parsing html of page [%s] - %s", page.URL, e.Error())
			}
			break
		}

		token := tokeniser.Token()

		if linkTagStart != nil && token.Type == html.TextToken {
			linkTagText = fmt.Sprintf("%s%s", linkTagText, token.Data)
		}

		// find <a> (link) tags and extract the link & text from them
		if token.DataAtom == atom.A {
			switch token.Type {

			case html.StartTagToken:
				if len(token.Attr) > 0 {
					linkTagStart = &token
				}

			case html.EndTagToken:
				if linkTagStart == nil {
					logger.Warnf("link end tag [%v] without start tag found in page [%s]", token, page.URL)
					continue
				}

				// get link from start tag token
				link := getLinkFromToken(linkTagStart)

				// improve link form if necessary
				link, e := FixLinkForm(page.URL, link)
				if e != nil {
					logger.Errorf("problem with link [%s] - %s", link, e)
					continue
				}

				// check if link is valid to be added to the link tree
				if IsValidLink(link, linkTagText, children) {
					children = append(children, NewPage(link, linkTagText, depth, page))
					logger.Infof("link found [%v] in page [%s]", link, page.URL)
				}

				linkTagStart = nil
				linkTagText = ""
			}
		}
	}
	return
}

func getLinkFromToken(token *html.Token) (link string) {
	for i := range token.Attr {
		if token.Attr[i].Key == "href" {
			link = strings.TrimSpace(token.Attr[i].Val)
		}
	}
	return link
}

// IsCrawlable decides whether to parse a page (i.e. crawl further)
func (page *Page) IsCrawlable(visitedURLs *ConcurrentMap, seenContent *ConcurrentMap) bool {
	config := crawlerConfig.Get()

	// check max depth has not yet been reached
	if page.Depth >= config.MaxDepth {
		logger.Infof("page [%s] not crawlable - max depth reached", page.URL)
		return false
	}

	urlHost, e := GetURLDomain(page.URL)
	if e != nil {
		logger.Errorf("unexpected error checking page is crawlable, could not parse URL - [%s]", e)
		return false
	}

	// check site is not in configured blacklist
	for _, blacklisted := range config.BlacklistedURLs {
		blacklistedHost, e := GetURLDomain(blacklisted)
		if e != nil {
			logger.Errorf("could not parse blacklist URL [%s] - [%s]", blacklisted, e)
			continue
		}
		if urlHost == blacklistedHost {
			logger.Infof("page [%s] not crawlable - site blacklisted", page.URL)
			return false
		}
	}

	// check url has not yet been crawled
	if visitedURLs.KeyExists(page.URLHash) {
		logger.Infof("page [%s] not crawlable - url already visited", page.URL)
		return false
	}

	// check if page content has already been seen, perhaps via a different URL
	if seenContent.KeyExists(page.ContentHash) {
		logger.Infof("page [%s] not crawlable - content already seen", page.URL)
		return false
	}

	return true
}

// PrintTree prints the site map to the console
func (page *Page) PrintTree() {
	if page == nil {
		return
	}

	if page.Depth == 0 {
		fmt.Print("\n\n")
	}

	indentSize := 60

	depthString := fmt.Sprintf("[%d] ", page.Depth)
	fmt.Printf(depthString+"%-*s", indentSize, page.URL)

	if page.Children != nil && len(page.Children) != 0 {
		for i, child := range page.Children {
			if i != 0 {
				fmt.Print("\n")
				fmt.Printf("%*s", (indentSize+len(depthString))*(page.Depth+1), "")
			}
			child.PrintTree()
		}
	}
}
