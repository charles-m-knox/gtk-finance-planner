.PHONY=build

BUILDDIR=build
VER=0.1.2
FILE=gtk-finance-planner
BIN=$(BUILDDIR)/$(FILE)-v$(VER)
OUT_BIN_DIR=~/.local/bin
UNAME=$(shell go env GOOS)
ARCH=$(shell go env GOARCH)
BUILD_ENV=CGO_ENABLED=1
BUILD_FLAGS=-ldflags="-w -s -buildid= -X constants.VERSION=$(VER)" -trimpath
GPG_SIGNING_KEY=$(shell git config --get user.signingkey)
FLATPAK_BUILD_DIR=$(BUILDDIR)/flatpak
FLATPAK_REPOSITORY=/mnt/flatpakrepo-charlesmknox
FLATPAK_MANIFEST=com.charlesmknox.gtk-finance-planner.yml

build-dev:
	$(BUILD_ENV) go build -v

mkbuilddir:
	mkdir -p $(BUILDDIR)

build-prod: mkbuilddir
	make build-$(UNAME)-$(ARCH)

test:
	go test -test.v -coverprofile=testcov.out ./... && \
	go tool cover -html=testcov.out

run:
	./$(BIN)

lint:
	golangci-lint run ./...

install:
	rsync -avP ./$(BIN)-$(UNAME)-$(ARCH) $(OUT_BIN_DIR)/$(FILE)
	chmod +x $(OUT_BIN_DIR)/$(FILE)

compress-prod: mkbuilddir
	rm -f $(BIN)-compressed
	upx --best -o ./$(BIN)-compressed $(BIN)

build-darwin-arm64: mkbuilddir
	$(BUILD_ENV) GOARCH=arm64 GOOS=darwin go build -v -o $(BIN)-darwin-arm64 $(BUILD_FLAGS)
	rm -f $(BIN)-darwin-arm64.xz
	xz -9 -e -T 12 -vv $(BIN)-darwin-arm64

build-darwin-amd64: mkbuilddir
	$(BUILD_ENV) GOARCH=amd64 GOOS=darwin go build -v -o $(BIN)-darwin-amd64 $(BUILD_FLAGS)
	rm -f $(BIN)-darwin-amd64.xz
	xz -9 -e -T 12 -vv $(BIN)-darwin-amd64

build-win-amd64: mkbuilddir
	$(BUILD_ENV) GOARCH=amd64 GOOS=windows go build -v -o $(BIN)-win-amd64-uncompressed $(BUILD_FLAGS)
	rm -f $(BIN)-win-amd64
	upx --best -o ./$(BIN)-win-amd64 $(BIN)-win-amd64-uncompressed

build-linux-arm64: mkbuilddir
	$(BUILD_ENV) GOARCH=arm64 GOOS=linux go build -v -o $(BIN)-linux-arm64-uncompressed $(BUILD_FLAGS)
	rm -f $(BIN)-linux-arm64
	upx --best -o ./$(BIN)-linux-arm64 $(BIN)-linux-arm64-uncompressed

build-linux-amd64: mkbuilddir
	$(BUILD_ENV) GOARCH=amd64 GOOS=linux go build -v -o $(BIN)-linux-amd64-uncompressed $(BUILD_FLAGS)
	rm -f $(BIN)-linux-amd64
	upx --best -o ./$(BIN)-linux-amd64 $(BIN)-linux-amd64-uncompressed

# as of 2024-08-02, building for arm64 doesn't seem to work.
# build-all: mkbuilddir build-linux-amd64 build-linux-arm64 build-win-amd64 build-mac-amd64 build-mac-arm64
build-all: mkbuilddir build-linux-amd64 build-win-amd64 build-mac-amd64 build-mac-arm64

delete-uncompressed:
	rm $(BUILDDIR)/*-uncompressed

delete-builds:
	rm $(BUILDDIR)/*

flatpak-build-test:
	mount --fake | grep -i $(FLATPAK_REPOSITORY)
	rm -rf $(FLATPAK_BUILD_DIR)
	mkdir -p $(FLATPAK_BUILD_DIR)
	flatpak --user install runtime/org.freedesktop.Sdk/x86_64/23.08
	flatpak --user install runtime/org.freedesktop.Platform/x86_64/23.08
	flatpak-builder --user --install --gpg-sign=$(GPG_SIGNING_KEY) $(FLATPAK_BUILD_DIR) $(FLATPAK_MANIFEST)

flatpak-release:
	mount --fake | grep -i $(FLATPAK_REPOSITORY)
	rm -rf $(FLATPAK_BUILD_DIR)
	mkdir -p $(FLATPAK_BUILD_DIR)
	flatpak --user install runtime/org.freedesktop.Sdk/x86_64/23.08
	flatpak --user install runtime/org.freedesktop.Platform/x86_64/23.08
	flatpak-builder --user --install --gpg-sign=$(GPG_SIGNING_KEY) --repo=$(FLATPAK_REPOSITORY) $(FLATPAK_BUILD_DIR) $(FLATPAK_MANIFEST)
