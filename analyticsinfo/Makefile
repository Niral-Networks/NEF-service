# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

# Path
MAIN_PATH=./nef.go
BIN_PATH=./bin/$(BINARY_NAME)
SRC_YAML_PATH=./config
SRC_BIN_PATH=./bin
DEST_YAML_PATH=../../install/etc/niralos
DEST_BIN_PATH=../../install/bin
YAML_FILE=nefcfg.yaml
BIN_FILE=niralos-nefd

# Binary name
BINARY_NAME=niralos-nefd

# Targets
.PHONY: all debug build clean

all: build

debug: GCFLAGS += -N -l
debug: all

build:
	$(GOBUILD) -gcflags "$(GCFLAGS)" -o $(BIN_PATH) $(MAIN_PATH)

clean:
	$(GOCLEAN)
	rm -f $(BIN_PATH)

copy-yaml:
	@if [ -f $(SRC_YAML_PATH)/$(YAML_FILE) ]; then \
                mkdir -p $(DEST_YAML_PATH); \
                cp $(SRC_YAML_PATH)/$(YAML_FILE) $(DEST_YAML_PATH); \
                echo "Copied $(YAML_FILE) to $(DEST_YAML_PATH)"; \
        else \
                echo "File $(SRC_YAML_PATH)/$(YAML_FILE) does not exist."; \
                exit 1; \
        fi
	@if [ -f $(SRC_BIN_PATH)/$(BIN_FILE) ]; then \
                mkdir -p $(DEST_BIN_PATH); \
                cp $(SRC_BIN_PATH)/$(BIN_FILE) $(DEST_BIN_PATH); \
                echo "Copied $(BIN_FILE) to $(DEST_BIN_PATH)"; \
        else \
                echo "File $(SRC_BIN_PATH)/$(BIN_FILE) does not exist."; \
                exit 1; \
        fi
