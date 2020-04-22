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
BUILD_VERSION := $(shell git describe --tags | grep -P -o '(?<=v)[0-9]+.[0-9]+.[0-9]')
DEB_DIR_BASE := debian
DEB_DIR := $(DEB_DIR_BASE)/goyammer_$(BUILD_VERSION)
DEB_PACKAGE := $(DEB_DIR_BASE)/goyammer_$(BUILD_VERSION).deb

define DEB_CONTROL
Package: goyammer
Version: $(BUILD_VERSION)
Section: misc
Priority: optional
Architecture: amd64
Depends: libblkid1 (>= 2.31), libc6 (>= 2.27), libffi6 (>= 3.2), libgdk-pixbuf2.0-0 (>= 2.36), libglib2.0-0 (>= 2.56), libmount1 (>= 2.31), libnotify4 (>= 0.7), libpcre3 (>= 2:8.39), libselinux1 (>= 2.7), libuuid1 (>= 2.31), zlib1g (>= 1:1.2)
Maintainer: Sebastian Bogan <sebogh@qibli.net>
Description: Notify about new Yammer messages.
 Poll the Yammer API and notify about new messages using libnotify.

endef
export DEB_CONTROL

all: goyammer

goyammer: $(GO_FILES) Makefile
	GO111MODULE=on CGO_ENABLED=1 GOOS=linux go build \
	-ldflags '-X main.buildVersion=$(BUILD_VERSION) -X main.buildGithash=$(BUILD_GITHASH)' \
	github.com/seboghpub/goyammer

# see: https://www.tldp.org/HOWTO/html_single/Debian-Binary-Package-Building-HOWTO/
$(DEB_DIR): goyammer Makefile
	mkdir -p $(DEB_DIR)/DEBIAN
	mkdir -p $(DEB_DIR)/usr/bin
	mkdir -p $(DEB_DIR)/usr/share/doc/goyammer
	mkdir -p $(DEB_DIR)/usr/share/man/man1/
	mkdir -p $(DEB_DIR)/usr/share/metainfo/
	@echo "$$DEB_CONTROL" > $(DEB_DIR)/DEBIAN/control
	cp goyammer $(DEB_DIR)/usr/bin
	strip --strip-unneeded --remove-section=.comment --remove-section=.note $(DEB_DIR)/usr/bin/goyammer
	cp copyright $(DEB_DIR)/usr/share/doc/goyammer
	cp changelog $(DEB_DIR)/usr/share/doc/goyammer
	gzip --best --no-name $(DEB_DIR)/usr/share/doc/goyammer/changelog
	pandoc goyammer.1.md -s -t man -o $(DEB_DIR)/usr/share/man/man1/goyammer.1
	gzip --best --no-name $(DEB_DIR)/usr/share/man/man1/goyammer.1
	cp appdata.xml $(DEB_DIR)/usr/share/metainfo/goyammer.appdata.xml

$(DEB_PACKAGE): $(DEB_DIR)
	fakeroot dpkg-deb --build $(DEB_DIR)
	lintian $@ || true

debbuild: tidy $(DEB_PACKAGE)


# use this to find out .deb dependencies
showdebs: goyammer
	ldd goyammer | grep -P -o "(?<=> ).+(?= \(0x)" | xargs dpkg -S | awk -F ':' '{print $$1}' | sort | uniq

clean:
	go clean github.com/seboghpub/goyammer
	rm -f *~
	rm -Rf $(DEB_DIR_BASE)

tidy: clean
	rm -f goyammer
	rm -f $(DEB_PACKAGE)


.PHONY: all clean tidy debbuild showdebs
