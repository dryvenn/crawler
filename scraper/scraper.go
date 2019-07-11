// Package scraper holds a webpage link scrapper
package scraper

import (
	"io"
	"net/http"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	log "github.com/sirupsen/logrus"
)

// Scraper is a web page scrapper that outputs links within them.
type Scraper struct{}

type urlScrapper struct {
	url string
}

func (s urlScrapper) logger() log.FieldLogger {
	return log.WithField("url", s.url)
}

func (s urlScrapper) extractLinks(data io.Reader) []string {
	log := s.logger()
	ret := make([]string, 0)
	tz := html.NewTokenizer(data)
	for {
		switch tz.Next() {
		case html.ErrorToken:
			if err := tz.Err(); err != io.EOF {
				log.WithError(err).Error("Error parsing data")
			}
			return ret
		case html.StartTagToken:
			token := tz.Token()
			// Is this an <a> tag?
			if token.DataAtom != atom.A {
				continue
			}
			// Find the href attribute's content.
			var link string
			for _, a := range token.Attr {
				if a.Key == "href" {
					link = a.Val
					break
				}
			}
			if link == "" {
				log.Errorf("Expected href attr inside token %v", token)
				continue
			}
			// Looking good, keep it!
			ret = append(ret, link)
		}
	}
}

// ScrapeLinks gets the page at the given url and returns all the links within it.
func (s Scraper) ScrapeLinks(url string) ([]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return urlScrapper{url: url}.extractLinks(res.Body), nil
}
