GOMOD = go mod
GOBUILD = go build -trimpath -v
GOTEST = go test -v -cover -race

ROOT = $(shell git rev-parse --show-toplevel)
BIN = dist/grinfo
CMD = "./cmd/grinfo"

.PHONY: $(BIN)
$(BIN):
	$(GOBUILD) -o $@ $(CMD)

.PHONY: test
test:
	$(GOTEST) ./...

.PHONY: init
init:
	$(GOMOD) tidy

.PHONY: vuln
vuln:
	go run golang.org/x/vuln/cmd/govulncheck ./...

.PHONY: vet
vet:
	go vet ./...
