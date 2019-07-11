mmdc := node_modules/.bin/mmdc

.PHONY: all
ifeq ($(URL),)
all: crawl
else
all: urlgraph
endif

OUT := $(shell echo $(URL) | sed 's@.*://@@')


crawl: ./cmd/crawl/crawl.go
	go build $^


.PHONY: urlgraph
urlgraph: $(OUT).svg
	firefox $<


$(OUT).svg: $(OUT).mmd $(mmdc)
	$(mmdc) -i $< -o $@


$(OUT).mmd: crawl
	./crawl -mermaid $(URL) > $@


$(mmdc):
	npm install mermaid.cli
