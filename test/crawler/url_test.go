package crawler

import (
	"fmt"
	"testing"
	"webcrawler/internal/crawler"
)

func TestGetURLDomain(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedString string
		expectedError  error
	}{
		{
			name:           "success_standard",
			url:            "https://www.google.com",
			expectedString: "www.google.com",
			expectedError:  nil,
		},
		{
			name:           "success_subdomain",
			url:            "https://images.google.com",
			expectedString: "images.google.com",
			expectedError:  nil,
		},
		{
			name:           "fail_incomplete",
			url:            "/about",
			expectedString: "",
			expectedError:  fmt.Errorf("could not derive URL domain, incomplete URL provided - [/about]"),
		},
		{
			name:           "fail_malformed",
			url:            "",
			expectedString: "",
			expectedError:  fmt.Errorf("could not derive URL domain, error parsing url [] - [parse \"\\x10\": net/url: invalid control character in URL]"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			domain, e := crawler.GetURLDomain(test.url)

			if domain != test.expectedString {
				t.Errorf("unexpected domain.\n- received: %s\n- expected %s", domain, test.expectedString)
			}
			if e == nil && test.expectedError != nil {
				t.Errorf("error is nil.\n- received: %s\n- expected %s", e, test.expectedError)
			}
			if e != nil && test.expectedError == nil {
				t.Errorf("unexpected error.\n- received: %s\n- expected %s", e, test.expectedError)
			}
			if e != nil && e.Error() != test.expectedError.Error() {
				t.Errorf("mismatched error message.\n- received: %s\n- expected %s", e, test.expectedError)
			}
		})
	}
}

func TestFixShortcutLink(t *testing.T) {
	tests := []struct {
		name           string
		parentURL      string
		link           string
		expectedString string
		expectedError  error
	}{
		{
			name:           "success_fix",
			parentURL:      "https://www.google.com",
			link:           "/search",
			expectedString: "https://www.google.com/search",
			expectedError:  nil,
		},
		{
			name:           "success_no-fix",
			parentURL:      "https://www.google.com",
			link:           "https://www.google.com/search",
			expectedString: "https://www.google.com/search",
			expectedError:  nil,
		},
		{
			name:           "fail_malformed-link",
			parentURL:      "https://www.google.com",
			link:           "https://www.google.com/",
			expectedString: "https://www.google.com/",
			expectedError:  fmt.Errorf("could not test link format, error parsing url [https://www.google.com/] - [parse \"https://www.google.com/\\x10\": net/url: invalid control character in URL]"),
		},
		{
			name:           "fail_malformed-parent",
			parentURL:      "https://www.google.com/",
			link:           "/search",
			expectedString: "/search",
			expectedError:  fmt.Errorf("could not fix link format, error parsing parent url [https://www.google.com/] - [parse \"https://www.google.com/\\x10\": net/url: invalid control character in URL]"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			domain, e := crawler.FixShortcutLink(test.parentURL, test.link)

			if domain != test.expectedString {
				t.Errorf("unexpected domain.\n- received: %s\n- expected %s", domain, test.expectedString)
			}
			if e == nil && test.expectedError != nil {
				t.Errorf("error is nil.\n- received: %s\n- expected %s", e, test.expectedError)
			}
			if e != nil && test.expectedError == nil {
				t.Errorf("unexpected error.\n- received: %s\n- expected %s", e, test.expectedError)
			}
			if e != nil && e.Error() != test.expectedError.Error() {
				t.Errorf("mismatched error message.\n- received: %s\n- expected %s", e, test.expectedError)
			}
		})
	}
}

func TestIsValidLink(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		linkText        string
		currentChildren []*crawler.Page
		expectedResult  bool
	}{
		{
			name:            "success_valid",
			url:             "https://www.google.com",
			linkText:        "Google",
			currentChildren: []*crawler.Page{},
			expectedResult:  true,
		},
		{
			name:            "success_invalid_exists",
			url:             "https://www.google.com",
			linkText:        "Google",
			currentChildren: []*crawler.Page{{URL: "https://www.google.com"}},
			expectedResult:  false,
		},
		{
			name:            "success_invalid_url",
			url:             "",
			linkText:        "Google",
			currentChildren: []*crawler.Page{},
			expectedResult:  false,
		},
		{
			name:            "success_invalid_linkText",
			url:             "https://www.google.com",
			linkText:        "",
			currentChildren: []*crawler.Page{},
			expectedResult:  false,
		},
		{
			name:            "success_nil_children",
			url:             "https://www.google.com",
			linkText:        "",
			currentChildren: nil,
			expectedResult:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			isValidLink := crawler.IsValidLink(test.url, test.linkText, test.currentChildren)

			if isValidLink != test.expectedResult {
				t.Errorf("unexpected result.\n- received: %t\n- expected %t", isValidLink, test.expectedResult)
			}
		})
	}
}

func TestTrimLinkVars(t *testing.T) {
	tests := []struct {
		name           string
		link           string
		expectedResult string
	}{
		{
			name:           "success_trim",
			link:           "https://www.google.com?a=b",
			expectedResult: "https://www.google.com",
		},
		{
			name:           "success_no_trim",
			link:           "https://www.google.com?a=b",
			expectedResult: "https://www.google.com",
		},
		{
			name:           "success_no_trim",
			link:           "/search",
			expectedResult: "/search",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			link := crawler.TrimLinkVars(test.link)

			if link != test.expectedResult {
				t.Errorf("unexpected result.\n- received: %s\n- expected %s", link, test.expectedResult)
			}
		})
	}
}
