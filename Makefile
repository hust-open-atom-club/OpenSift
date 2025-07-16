# Makefile
# Go commands
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOFMT = $(GOCMD) fmt

# Binary directory
BIN_DIR=./bin

# foreach dir in cmd/*/main.go, set * as the target
# APPS = $(patsubst cmd/%/main.go,%,$(wildcard cmd/**/*/main.go))
APP_ENTRIES = $(wildcard cmd/*/main.go) $(wildcard cmd/*/*/main.go)
APPS_ALL = $(patsubst cmd/%/main.go,%,$(APP_ENTRIES))
APPS = $(filter-out archives/%,$(APPS_ALL))

# $(info $(APPS))

# Default target
all: $(APPS)

# all app targets
$(APPS): %:
	$(GOBUILD) -o $(BIN_DIR)/$@ github.com/HUSTSecLab/OpenSift/cmd/$@

# # all binaries
# $(BIN_DIR)/%: cmd/%

fmt:
	$(GOFMT) ./...

clean:
	$(GOCLEAN)
	rm -rf $(BIN_DIR)

test:
	$(GOTEST) -v ./...

.PHONY: all build $(APPS) clean test fmt