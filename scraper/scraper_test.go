package scraper

import (
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

func sortAndCompare(a []string, b []string) bool {
	sort.Strings(a)
	sort.Strings(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestURLScrapper_extractLinks(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	t.Run("test finding links in page", func(t *testing.T) {
		data := []struct {
			document string
			links    []string
		}{
			{
				document: `<a href="https://google.com">google</a>`,
				links: []string{
					"https://google.com",
				},
			},
			{
				document: `
			<head><a href="https://google.com">google</a></head>
			<body>
				<a href="https://amazon.com">amazon</a>
				<div>
					<a href="https://facebook.com">facebook</a>
				</div>
				<div>
					<p>
						<a href="https://apple.com">apple</a>
					</p>
			</body>
			`,
				links: []string{
					"https://google.com",
					"https://amazon.com",
					"https://facebook.com",
					"https://apple.com",
				},
			},
			{
				document: ``,
				links:    []string{},
			},
			{
				document: `<a>google</a>`,
				links:    []string{},
			},
			{
				document: `<a href="">`,
				links:    []string{},
			},
			{
				document: `<a href="google">`,
				links:    []string{"google"},
			},
		}

		for _, d := range data {
			links := urlScrapper{url: ""}.extractLinks(strings.NewReader(d.document))
			if !sortAndCompare(links, d.links) {
				t.Errorf("%v != %v", links, d.links)
			}
		}
	})
}

func TestScrapper_ScrapeLinks(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	t.Run("test scraping real pages", func(t *testing.T) {
		data := []struct {
			url  string
			fake bool
		}{
			{
				url:  "https://google.com",
				fake: false,
			},
			{
				url:  "https://example.example",
				fake: true,
			},
		}

		for _, d := range data {
			links, err := Scraper{}.ScrapeLinks(d.url)
			if d.fake != (err != nil) || d.fake != (len(links) == 0) {
				t.Errorf("fake=%v url=%v links=%v err=%v", d.fake, d.url, links, err)
			}
		}
	})
}
