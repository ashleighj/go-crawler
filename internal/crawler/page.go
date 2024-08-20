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

func (page *Page) AddCrawlData(contentString string, children []*Page) {
	page.RawContent = contentString
	page.ContentHash = util.Hash(contentString)
	page.Children = children
}

// GetChildren finds the links in a page, uses them to contruct new
// Page structs relating to the parent, and returns those new Page structs
func (page *Page) GetChildren(pageBody io.ReadCloser, depth int) (children []*Page) {
	logger.Infof("parsing page at [%s], finding children links", page.URL)
	// split page into tokens
	tokeniser := html.NewTokenizer(pageBody)

	var linkTagStart *html.Token
	var linkTagText string

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

			case html.SelfClosingTagToken:
				logger.Infof("SelfClosingTagToken found - [%v]", token)

			case html.EndTagToken:
				if linkTagStart == nil {
					logger.Warnf("link end tag [%v] without start tag found in page [%s]", token, page.URL)
					continue
				}

				link := ""
				for i := range linkTagStart.Attr {
					if linkTagStart.Attr[i].Key == "href" {
						link = strings.TrimSpace(linkTagStart.Attr[i].Val)
					}
				}

				link = strings.ReplaceAll(FixShortcutLink(page.URL, link), " ", "")

				if IsValidLink(link, linkTagText, depth, children) {
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

// IsCrawlable decides whether to parse a page (i.e. crawl further)
func (page *Page) IsCrawlable(visitedURLs *ConcurrentMap, seenContent *ConcurrentMap) bool {
	config := crawlerConfig.Get()
	if page.Depth >= config.MaxDepth {
		logger.Infof("page [%s] not crawlable - max depth reached", page.URL)
		return false
	}

	if visitedURLs.KeyExists(page.URLHash) {
		logger.Infof("page [%s] not crawlable - url already visited", page.URL)
		return false
	}
	if seenContent.KeyExists(page.ContentHash) {
		logger.Infof("page [%s] not crawlable - content already seen", page.URL)
		return false
	}

	// TODO check for blacklisted sites

	return true
}

// TODO
// IsAlreadyVisited checks whether a page has already been crawled
func (page *Page) IsAlreadyVisited() (visited bool, e error) {
	return
}

// PrintTree prints the site map to the console
func (page *Page) PrintTree() {
	if page == nil {
		return
	}

	if page.Depth == 0 {
		fmt.Print("\n\n")
	}

	indent := 60

	fmt.Printf("[%d] %-*s", page.Depth, indent, page.URL)

	if page.Children != nil && len(page.Children) != 0 {
		for i, child := range page.Children {
			if i != 0 {
				fmt.Print("\n")
				fmt.Printf("%*s", indent*(page.Depth+1), "")
			}
			child.PrintTree()
		}
	}
}
