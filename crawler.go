// Package crawler implements a simple webpage crawler that aims to output
// relationship between pages within a single subdomain.
package crawler

import (
	"fmt"
	"net/url"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/dryvenn/crawler/scraper"
)

// Page is a webpage full of links.
type Page struct {
	URL   string
	Links []string
}

type linksScraper interface {
	ScrapeLinks(string) ([]string, error)
}

type crawler struct {
	baseURL *url.URL
	pages   chan Page
	linksScraper
}

// Filter returns a new page whose links has been filtered:
// - invalid links are taken out
// - duplicates are taken out
// - absolute links with a different host than the URL are taken out
// - relative links are made absolute
// - the query and fragment of the links are removed
func (p Page) filterLinks() Page {
	pageURL, err := url.Parse(p.URL)
	if err != nil {
		// This should never happen, a Page's URL should always be valid.
		// In this case, let's give up: it is impossible to do filtering.
		log.WithError(err).WithField("page", p).Error("Page as invalid URL")
		return p
	}
	filtered := make(map[string]struct{})
	for _, l := range p.Links {
		// Validate the link as URL.
		linkURL, err := url.Parse(l)
		if err != nil {
			log.WithError(err).WithField("url", l).Debug("Invalid page link")
			continue
		}
		if linkURL.IsAbs() {
			// For absolute URLs, check the Host is the same as the page's.
			// Allow for a different scheme though.
			if linkURL.Host != pageURL.Host {
				continue
			}
		} else {
			// For relative URLs, fill the scheme and host fields.
			linkURL.Scheme = pageURL.Scheme
			linkURL.Host = pageURL.Host
		}
		// Either way, get rid of the query.
		linkURL.RawQuery = ""
		linkURL.Fragment = ""
		// Only keep basic info.
		filtered[linkURL.String()] = struct{}{}
	}
	// Return the page with filtered links.
	links := make([]string, 0, len(filtered))
	for k := range filtered {
		links = append(links, k)
	}
	return Page{
		URL:   p.URL,
		Links: links,
	}
}

func (c crawler) start() {
	// Keep track of what has been scraped before.
	scrapedLinksRecord := make(map[string]struct{})
	// Collect scrape output from here.
	scrapedPages := make(chan Page, 100)

	// This here is the stop condition: wait for all results to come in,
	// and when they do signal it to the main loop.
	// Note: we can't start the stop condition wait has not work has yet
	// been stacked up (see below).
	var wg sync.WaitGroup
	triggerWait := func() {
		wg.Wait()
		close(scrapedPages)
	}

	scrapeURL := func(u string) {
		// Mark this record as scraped.
		scrapedLinksRecord[u] = struct{}{}
		// Increment the waitgroup to signal pending result.
		wg.Add(1)
		go func() {
			log := log.WithField("url", u)
			log.Debug("Scrapping URL")
			links, err := c.ScrapeLinks(u)
			if err != nil {
				log.WithError(err).Error("Scraping URL")
				// Decrement the waitgroup: there will be no result for this one.
				wg.Done()
				return
			}
			scrapedPages <- Page{URL: u, Links: links}.filterLinks()
		}()
	}

	// Now trigger the first scrape to increment the waitgroup, and then
	// it'll be safe to activate the stop condition.
	scrapeURL(c.baseURL.String())
	go triggerWait()

	for {
		page, ok := <-scrapedPages
		if !ok {
			// Crawling is finished, let's exit!
			log.WithField("url", c.baseURL).Info("Finished crawling")
			close(c.pages)
			return
		}
		// Trigger new scraps if necessary.
		for _, l := range page.Links {
			// Check if this link has been scraped before.
			_, ok := scrapedLinksRecord[l]
			if ok {
				// Been there, done that.
				continue
			}
			// Not scraped yet, just do it, do it!
			scrapeURL(l)
		}
		// Send the result for this page away.
		c.pages <- page
		// A result has been retrieved and its processing has ended.
		// Note that it is important to decrement *after* links within
		// this results have had the opportunity to increment.
		wg.Done()
	}
}

// Crawl starts crawling the given URL, unless an error is returned first.
// The crawling stopes when the pages chan is closed.
func Crawl(s string) (chan Page, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("invalid url for crawling: %v", err)
	}

	c := crawler{
		baseURL:      u,
		pages:        make(chan Page, 100),
		linksScraper: scraper.Scraper{},
	}

	go c.start()

	return c.pages, nil
}
