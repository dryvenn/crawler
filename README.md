# Simple subdomain crawler

Given an URL, crawl the webpages of its subdomain.

## How to compile and run

> Tested with Go1.12, you probably need at least Go1.11 for the modules support.

```
# Pulling from Github:
go get -u https://github.com/dryvenn/crawler/cmd/...
crawl

# Or from a local clone:
go build ./cmd/crawl/
./crawl
```
