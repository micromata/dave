.PHONY: fmt check

SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

fmt:
	@gofmt -s -l -w $(SRC)

check:
	@for i in $$(go list ./... | grep -v /vendor/); do golint $${i}; done
	@go tool vet ${SRC}
