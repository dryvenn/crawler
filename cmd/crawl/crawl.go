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

func simpleOutput(pages chan crawler.Page) {
	for page := range pages {
		fmt.Printf("%s: %s\n", page.URL, strings.Join(page.Links, ", "))
	}
}

func mermaidOutput(pages chan crawler.Page) {
	idCounter := 0
	idDirectory := make(map[string]int)
	getID := func(u string) int {
		if id, ok := idDirectory[u]; ok {
			return id
		} else {
			id = idCounter
			idCounter += 1
			idDirectory[u] = id
			fmt.Println(fmt.Sprintf(`    id%d["%s"]`, id, u))
			return id
		}
	}
	fmt.Println(`graph TD`)
	for page := range pages {
		for _, link := range page.Links {
			fmt.Println(fmt.Sprintf(`    id%d --> id%d`, getID(page.URL), getID(link)))
		}
	}
}

func main() {
	debug := flag.Bool("debug", false, "Whether to enable logs")
	mermaid := flag.Bool("mermaid", false, "Change the output to be a mermaid graph")
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

	if *mermaid {
		mermaidOutput(pages)
	} else {
		simpleOutput(pages)
	}
}
