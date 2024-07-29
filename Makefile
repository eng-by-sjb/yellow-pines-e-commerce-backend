PROJECT_NAME = yellow_pines
SRC_DIR = ./
TEST_DIR = $(SRC_DIR)...
GO = @go

# GO commands
GO_BUILD = $(GO) build
GO_TEST = $(GO) test
GO_TIDY = $(GO) mod tidy

# Target

build: 
	$(GO_BUILD) -o bin/$(PROJECT_NAME) -v cmd/main.go

run: build
	bin/$(PROJECT_NAME)
	

test:
	$(GO_TEST) -v $(TEST_DIR)

tidy:
	$(GO_TIDY)