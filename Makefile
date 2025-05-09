.PHONY=build

BUILDDIR=build
VER=0.1.5
FILE=gtk-finance-planner
BIN=$(BUILDDIR)/$(FILE)-v$(VER)
OUT_BIN_DIR=~/.local/bin
UNAME=$(shell go env GOOS)
ARCH=$(shell go env GOARCH)
BUILD_ENV=CGO_ENABLED=1
BUILD_FLAGS=-ldflags="-w -s -buildid= -X constants.VERSION=$(VER)" -trimpath
GPG_SIGNING_KEY=$(shell git config --get user.signingkey)
FLATPAK_BUILD_DIR=$(BUILDDIR)/flatpak
FLATPAK_MANIFEST=com.charlesmknox.gtk-finance-planner.yml
FLATPAK_REPO_BASE_DIR=flatpakrepo
FLATPAK_REPO_DIR=$(FLATPAK_REPO_BASE_DIR)/repo
FLATPAK_REPO_GIT_BRANCH=flatpakrepo
FLATPAK_REPO_GIT_ORPHAN_BRANCH=flatpakrepo-tmp
FLATPAK_REPO_TMP_DIR=__TMP__FLATPAK__DIR__
FLATPAK_REPO_GHPAGES_BASE_DIR=docs
FLATPAK_REPO_GHPAGES_REPO_DIR=$(FLATPAK_REPO_GHPAGES_BASE_DIR)/repo
GIT_REMOTE=origin
GIT_MAIN_BRANCH=main

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

# warning: this will install the flatpak locally, and you might have to
# uninstall it & remove the (local) remote if you want to actually install the
# published flatpak
flatpak-build:
	rm -rf $(FLATPAK_BUILD_DIR) $(FLATPAK_REPO_DIR)
	mkdir -p $(FLATPAK_BUILD_DIR) $(FLATPAK_REPO_DIR)
	flatpak --user install runtime/org.freedesktop.Sdk/x86_64/23.08
	flatpak --user install runtime/org.freedesktop.Platform/x86_64/23.08
	flatpak-builder --user --install --gpg-sign=$(GPG_SIGNING_KEY) $(FLATPAK_BUILD_DIR) $(FLATPAK_MANIFEST)
	flatpak build-export --gpg-sign=$(GPG_SIGNING_KEY) $(FLATPAK_REPO_DIR) $(FLATPAK_BUILD_DIR)
	flatpak build-update-repo --gpg-sign=$(GPG_SIGNING_KEY) $(FLATPAK_REPO_DIR)

# warning: this is a dangerous operation that does a force-push and can delete
# local files
flatpak-publish: flatpak-build
	mv $(FLATPAK_REPO_DIR) $(FLATPAK_REPO_TMP_DIR)
	git checkout $(FLATPAK_REPO_GIT_BRANCH)
	rm -rf $(FLATPAK_REPO_GHPAGES_REPO_DIR)
	mv $(FLATPAK_REPO_TMP_DIR) $(FLATPAK_REPO_GHPAGES_REPO_DIR)
	! git diff --quiet || exit 1
	-git branch -D $(FLATPAK_REPO_GIT_ORPHAN_BRANCH)
	git checkout --orphan $(FLATPAK_REPO_GIT_ORPHAN_BRANCH)
	git add -A
	git commit -S -m "flatpakrepo build"
	-git branch -D $(FLATPAK_REPO_GIT_BRANCH)
	git checkout -b $(FLATPAK_REPO_GIT_BRANCH)
	-git branch -D $(FLATPAK_REPO_GIT_ORPHAN_BRANCH)
	git push -f $(GIT_REMOTE) $(FLATPAK_REPO_GIT_BRANCH)
	git checkout $(GIT_MAIN_BRANCH)
