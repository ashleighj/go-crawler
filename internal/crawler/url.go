package crawler

import (
	"fmt"
	"net/url"
	"strings"
	crawlerConfig "webcrawler/config/crawler"
	logger "webcrawler/logger"
)

// GetURLDomain returns the subdomain & domain portion of a given URL
func GetURLDomain(urlString string) (domain string, e error) {
	parsed, e := url.Parse(urlString)

	if e != nil {
		e = fmt.Errorf("could not derive URL domain, error parsing url [%s] - [%s]", urlString, e)
		logger.Error(e)
		return
	}

	if parsed.Host == "" {
		e = fmt.Errorf("could not derive URL domain, incomplete URL provided - [%s]", urlString)
		logger.Error(e)
		return
	}

	return parsed.Host, nil
}

// FixShortcutLink checks if a link is relative (i.e. not fully-formed) and, if so, adds the schema and domain of its parent page
func FixShortcutLink(parentURL string, link string) (fixedLink string, e error) {
	parsed, e := url.Parse(link)
	if e != nil {
		e = fmt.Errorf("could not test link format, error parsing url [%s] - [%s]", link, e)
		logger.Error(e)
		return link, e
	}

	if parsed.Scheme != "" && parsed.Host != "" {
		return link, nil
	}

	logger.Infof("found relative link [%s] to fix  in page [%s]", link, parentURL)

	parent, e := url.Parse(parentURL)
	if e != nil {
		e = fmt.Errorf("could not fix link format, error parsing parent url [%s] - [%s]", parentURL, e)
		logger.Error(e)
		return link, e
	}

	if parsed.Host == "" {
		link = strings.TrimPrefix(link, ".")
		link = strings.TrimPrefix(link, "/")
	}

	fixedLink = fmt.Sprintf("%s://%s/%s", parent.Scheme, parent.Host, link)
	logger.Infof("fixed relative link [%s] in page [%s]", fixedLink, parentURL)
	return
}

// IsValidLink decides if a link is valid to be added to the page tree
func IsValidLink(url string, linkText string, currentChildren []*Page) (isValid bool) {
	if url == "" {
		return false
	}
	if linkText == "" {
		return false
	}

	// check if link already added to parent page children
	for _, page := range currentChildren {
		if page.URL == url {
			return false
		}
	}

	// get config to check for blacklisted sites
	config := crawlerConfig.Get()

	// check if blacklisted
	for _, ignoreable := range config.IgnoreIfContains {
		if strings.Contains(strings.ToLower(url), strings.ToLower(ignoreable)) {
			return false
		}
	}

	return true
}

// TrimLinkVars returns a URL as just a combination of its schema, domain and path, removing any query params
func TrimLinkVars(link string) string {
	return strings.Split(link, "?")[0]
}
