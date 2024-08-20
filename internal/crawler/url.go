package crawler

import (
	"fmt"
	"net/url"
	"strings"
	crawlerConfig "webcrawler/config/crawler"
	logger "webcrawler/logger"
)

func GetURLDomain(urlString string) (domain string, e error) {
	parsed, e := url.Parse(urlString)
	if e != nil {
		return
	}

	return parsed.Host, nil
}

func FixShortcutLink(parentURL string, link string) (fixedLink string) {
	parsed, e := url.Parse(link)
	if e != nil {
		logger.Error(e)
		return link
	}
	if parsed.Scheme != "" && parsed.Host != "" {
		return link
	}
	logger.Infof("found relative link [%s] to fix  in page [%s]", link, parentURL)

	if parsed.Host == "" {
		link = strings.TrimPrefix(link, ".")
		link = strings.TrimPrefix(link, "/")
	}

	parent, e := url.Parse(parentURL)
	if e != nil {
		logger.Error(e)
		return link
	}
	fixedLink = fmt.Sprintf("%s://%s/%s", parent.Scheme, parent.Host, link)
	logger.Infof("fixed relative link [%s] in page [%s]", fixedLink, parentURL)
	return
}

// IsValidLink decides if a link is valid to be added to the page tree
func IsValidLink(
	url string, text string, depth int, currentChildList []*Page) (isValid bool) {

	config := crawlerConfig.Get()

	for _, page := range currentChildList {
		if page.URL == url {
			return false
		}
	}

	if len(url) == 0 {
		return false
	}

	if len(text) == 0 {
		return false
	}

	for _, ignoreable := range config.IgnoreIfContains {
		if strings.Contains(strings.ToLower(url), strings.ToLower(ignoreable)) {
			return false
		}
	}

	// if depth >= config.MaxDepth {
	// 	return false
	// }

	return true
}

func TrimLinkVars(link string) string {
	return strings.Split(link, "?")[0]
}
