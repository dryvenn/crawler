// Package main is a basic crawler CLI.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"flag"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/dryvenn/crawler"
)

func usage() {
	fmt.Printf(
`Usage:
	> %s <url>

Crawl this URL for its subdomain only and display the results as a list of strings.

`, path.Base(os.Args[0]))
	flag.PrintDefaults()
		os.Exit(1)
}

func main() {
	debug := flag.Bool("debug", false, "Whether to enable logs")
	flag.Parse()

	if !*debug {
		log.SetOutput(ioutil.Discard)
	}

	if flag.NArg() != 1 {
		usage()
	}

	pages, err := crawler.Crawl(flag.Arg(0))
	if err != nil {
		fmt.Printf("Error starting crawling: %v\n", err)
		os.Exit(2)
	}

	for page := range pages {
		fmt.Printf("%s: %s\n", page.URL, strings.Join(page.Links, ", "))
	}
}
