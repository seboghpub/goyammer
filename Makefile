SHELL = /bin/bash

# Just to be sure, add the path of the binary-based go installation.
PATH := /usr/local/go/bin:$(PATH)

# Using the (above extended) path, query the GOPATH (i.e. the user's go path).
GOPATH := $(shell env PATH=$(PATH) go env GOPATH)

# Add $GOPATH/bin to path
PATH := $(GOPATH)/bin:$(PATH)

CWD := $(shell pwd)

# Collect all golang files in the current directory.
GO_FILES := $(wildcard internal/*.go main.go)

BUILD_GITHASH := $(shell git rev-parse HEAD)
BUILD_VERSION := 0.1.0
DEB_VERSION   := $(BUILD_VERSION)
DEB_NAME      := goyammer
DEB_DIR_BASE  := debian
DEB_DIR       := $(DEB_DIR_BASE)/$(DEB_NAME)_$(DEB_VERSION)

define DEB_CONTROL
Package: $(DEB_NAME)
Version: $(DEB_VERSION)
Section: misc
Priority: optional
Architecture: amd64
Depends: libnotify-dev (>= 0.7)
Maintainer: Sebastian Bogan <sebogh@qibli.net>
Description: Notify about new Yammer messages.
 Poll the Yammer API and notify about new messages using libnotify.

endef
export DEB_CONTROL

all: goyammer

vendor-sync:
	go mod vendor

goyammer: $(GO_FILES) Makefile vendor-sync
	CGO_ENABLED=1 GOOS=linux go build \
		-mod vendor \
		-ldflags '-X main.buildVersion=$(BUILD_VERSION) -X main.buildGithash=$(BUILD_GITHASH)' \
		github.com/seboghpub/goyammer

debdir: $(DEB_DIR)

$(DEB_DIR): clean goyammer Makefile
	mkdir -p $(DEB_DIR)/usr/bin
	cp goyammer $(DEB_DIR)/usr/bin

	# discard symbols and other data from object files
	strip --strip-unneeded --remove-section=.comment --remove-section=.note $(DEB_DIR)/usr/bin/goyammer

	mkdir $(DEB_DIR)/DEBIAN
	@echo "$$DEB_CONTROL" > $(DEB_DIR)/DEBIAN/control

	mkdir -p $(DEB_DIR)/usr/share/doc/goyammer
	cp copyright $(DEB_DIR)/usr/share/doc/goyammer
	cp changelog $(DEB_DIR)/usr/share/doc/goyammer
	gzip --best --no-name $(DEB_DIR)/usr/share/doc/goyammer/changelog


debbuild: $(DEB_DIR)
	fakeroot dpkg-deb --build $(DEB_DIR)

# Remove object files (if any).
clean:
	go clean github.com/seboghpub/goyammer
	rm -f *~
	rm -Rf $(DEB_DIR_BASE)

# Remove all intermediate files.
tidy: clean
	rm -f goyammer
	rm -Rf vendor


.PHONY: all clean tidy debdir

