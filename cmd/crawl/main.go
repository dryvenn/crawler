// Package main is a basic crawler CLI.
package main

import (
	"fmt"
	"github.com/dryvenn/crawler"
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetOutput(ioutil.Discard)
	if len(os.Args) != 2 {
		fmt.Printf(`
Usage: %s <url>

Crawl this URL for its subdomain only and display the results as a list of strings.
`, path.Base(os.Args[0]))
		os.Exit(1)
	}

	pages, err := crawler.Crawl(os.Args[1])
	if err != nil {
		fmt.Printf("Error starting crawling: %v\n", err)
		os.Exit(2)
	}

	for page := range pages {
		fmt.Printf("%s: %s\n", page.URL, strings.Join(page.Links, ", "))
	}
}
