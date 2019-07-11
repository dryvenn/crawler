package crawler

import (
	"errors"
	"io/ioutil"
	"net/url"
	"testing"

	log "github.com/sirupsen/logrus"
)

func stringToURL(t *testing.T, s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func stringSlicesEqualUnordered(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
NextValue:
	for i := range a {
		for j := range b {
			if a[i] == b[j] {
				continue NextValue
			}
		}
		return false
	}
	return true
}

func stringSlicesMapsEqual(a map[string][]string, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !stringSlicesEqualUnordered(a[k], b[k]) {
			return false
		}
	}
	return true
}

func TestPage_filterLinks(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	t.Run("test links are filtered according to base URL", func(t *testing.T) {
		data := []struct {
			baseURL string
			input   []string
			output  []string
		}{
			{
				baseURL: "https://example.com",
				input: []string{
					"https://example.com",
				},
				output: []string{
					"https://example.com",
				},
			},
			{
				baseURL: "https://example.com",
				input: []string{
					"/resources",
					"/",
					"data",
				},
				output: []string{
					"https://example.com/resources",
					"https://example.com/",
					"https://example.com/data",
				},
			},
			{
				baseURL: "https://example.com",
				input: []string{
					"https://example.com",
					"https://example.com/",
					"https://example.com//",
				},
				output: []string{
					"https://example.com",
					"https://example.com/",
					"https://example.com//",
				},
			},
			{
				baseURL: "https://example.com",
				input: []string{
					"https://example.com?hello=world",
					"https://example.com#L235",
				},
				output: []string{
					"https://example.com",
				},
			},
			{
				baseURL: "https://another.example.com",
				input: []string{
					"https://example.com",
					"https://another.example.com",
					"https://somethimg.com",
				},
				output: []string{
					"https://another.example.com",
				},
			},
		}
		for _, d := range data {
			res := Page{
				URL:   d.baseURL,
				Links: d.input,
			}.filterLinks().Links
			if !stringSlicesEqualUnordered(res, d.output) {
				t.Errorf("in=%v ref=%v res=%v", d.input, d.output, res)
			}
		}
	})
}

type staticLinksScraper map[string][]string

func (s staticLinksScraper) ScrapeLinks(u string) ([]string, error) {
	links, ok := s[u]
	if !ok {
		return nil, errors.New("unknown URL")
	}
	return links, nil
}

func TestCrawler_start(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	data := []struct {
		test  string
		ref   map[string][]string
		crawl crawler
	}{
		{
			test: "test that an empty base page is valid",
			ref: map[string][]string{
				"https://example.com": []string{},
			},
			crawl: crawler{
				baseURL: stringToURL(t, "https://example.com"),
				linksScraper: staticLinksScraper{
					"https://example.com": []string{},
				},
				pages: make(chan Page, 1),
			},
		},
		{
			test: "test multiple levels crawling works",
			ref: map[string][]string{
				"https://example.com": []string{
					"https://example.com/foo",
					"https://example.com/bar",
					"https://example.com/egg",
				},
				"https://example.com/foo": []string{},
				"https://example.com/bar": []string{},
				"https://example.com/egg": []string{
					"https://example.com/spam",
				},
				"https://example.com/spam": []string{},
			},
			crawl: crawler{
				baseURL: stringToURL(t, "https://example.com"),
				linksScraper: staticLinksScraper{
					"https://example.com": []string{
						"https://example.com/foo",
						"https://example.com/bar",
						"https://example.com/egg",
					},
					"https://example.com/foo": []string{},
					"https://example.com/bar": []string{},
					"https://example.com/egg": []string{
						"https://example.com/spam",
					},
					"https://example.com/spam": []string{},
				},
				pages: make(chan Page, 1),
			},
		},
		{
			test: "test that unscrapable pages are ignored",
			ref: map[string][]string{
				"https://example.com": []string{
					"https://example.com/foo",
					"https://example.com/bar",
					"https://example.com/egg",
				},
				"https://example.com/egg": []string{
					"https://example.com/spam",
				},
			},
			crawl: crawler{
				baseURL: stringToURL(t, "https://example.com"),
				linksScraper: staticLinksScraper{
					"https://example.com": []string{
						"https://example.com/foo",
						"https://example.com/bar",
						"https://example.com/egg",
					},
					"https://example.com/egg": []string{
						"https://example.com/spam",
					},
				},
				pages: make(chan Page, 1),
			},
		},
		{
			test: "test that invalid domains are ignored",
			ref: map[string][]string{
				"https://egg.example.com": []string{
					"https://egg.example.com/spam",
				},
			},
			crawl: crawler{
				baseURL: stringToURL(t, "https://egg.example.com"),
				linksScraper: staticLinksScraper{
					"https://egg.example.com": []string{
						"https://egg.example.com/spam",
						"https://foo.example.com",
						"https://example.com",
						"https://google.com",
					},
				},
				pages: make(chan Page, 1),
			},
		},
		{
			test: "test that loops do not hang",
			ref: map[string][]string{
				"https://egg.example.com": []string{
					"https://egg.example.com",
					"https://egg.example.com/spam",
				},
			},
			crawl: crawler{
				baseURL: stringToURL(t, "https://egg.example.com"),
				linksScraper: staticLinksScraper{
					"https://egg.example.com": []string{
						"https://egg.example.com",
						"https://egg.example.com/spam",
					},
				},
				pages: make(chan Page, 1),
			},
		},
	}
	for _, d := range data {
		t.Run(d.test, func(t *testing.T) {
			res := make(map[string][]string)
			go d.crawl.start()
			for p := range d.crawl.pages {
				res[p.URL] = p.Links
			}
			if !stringSlicesMapsEqual(res, d.ref) {
				t.Errorf("ref=%v res=%v", d.ref, res)
			}
		})
	}
}
