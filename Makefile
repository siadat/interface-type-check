.PHONY: dependencies test build

GOFORK = $(PWD)/go
CHECKER = ./cmd/interface-type-check

test: dependencies
	cd $(CHECKER) && GOROOT=$(GOFORK) $(GOFORK)/bin/go test -count=1 .

build: dependencies
	cd $(CHECKER) && GOOS=darwin GOARCH=amd64 GOROOT=$(GOFORK) $(GOFORK)/bin/go build -o interface-type-check .
	tar cvzf interface-type-check.darwin-amd64.tar.gz $(CHECKER)/interface-type-check
	rm $(CHECKER)/interface-type-check

	cd $(CHECKER) && GOOS=linux GOARCH=amd64 GOROOT=$(GOFORK) $(GOFORK)/bin/go build -o interface-type-check .
	tar cvzf interface-type-check.linux-amd64.tar.gz $(CHECKER)/interface-type-check
	rm $(CHECKER)/interface-type-check

	cd $(CHECKER) && GOOS=windows GOARCH=amd64 GOROOT=$(GOFORK) $(GOFORK)/bin/go build -o interface-type-check .
	tar cvzf interface-type-check.windows-amd64.tar.gz $(CHECKER)/interface-type-check
	rm $(CHECKER)/interface-type-check

dependencies: $(GOFORK)/bin/go
	cd $(CHECKER) && GOROOT=$(GOFORK) $(GOFORK)/bin/go get .

$(GOFORK)/bin/go:
	cd $(GOFORK)/src && ./make.bash
