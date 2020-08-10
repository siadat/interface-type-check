.PHONY: dependencies test build

GOROOT = $(PWD)/_go

test: dependencies
	GOROOT=$(GOROOT) $(GOROOT)/bin/go test -count=1 .

build: dependencies
	GOOS=darwin GOARCH=amd64 GOROOT=$(GOROOT) $(GOROOT)/bin/go build -o interface-type-check .
	tar cvzf interface-type-check.darwin-amd64.tar.gz interface-type-check
	rm interface-type-check

	GOOS=linux GOARCH=amd64 GOROOT=$(GOROOT) $(GOROOT)/bin/go build -o interface-type-check .
	tar cvzf interface-type-check.linux-amd64.tar.gz interface-type-check
	rm interface-type-check

dependencies:
	git submodule add -b interface-type-check --depth 1 https://github.com/siadat/go.git _go
	GOROOT=$(GOROOT) $(GOROOT)/bin/go get .
