SHELL = /bin/bash

# Just to be sure, add the path of the binary-based go installation.
PATH := /usr/local/go/bin:$(PATH)

# Using the (above extended) path, query the GOPATH (i.e. the user's go path).
GOPATH := $(shell env PATH=$(PATH) go env GOPATH)

# Add $GOPATH/bin to path
PATH := $(GOPATH)/bin:$(PATH)

CWD := $(shell pwd)

# Collect all golang files in the current directory.
GO_FILES := $(wildcard internal/*.go icon/*.go main.go)

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
Depends: libappindicator3-1 (>=12.10), libatk1.0-0 (>=2.28), libatk-bridge2.0-0 (>=2.26), libatspi2.0-0 (>=2.28), libblkid1 (>=2.31), libbsd0 (>=0.8), libc6 (>=2.27), libcairo2 (>=1.15), libcairo-gobject2 (>=1.15), libdatrie1 (>=0.2), libdbus-1-3 (>=1.12), libdbusmenu-glib4 (>=16.04), libdbusmenu-gtk3-4 (>=16.04), libepoxy0 (>=1.4), libexpat1 (>=2.2), libffi6 (>=3.2), libfontconfig1 (>=2.12), libfreetype6 (>=2.8), libgcrypt20 (>=1.8), libgdk-pixbuf2.0-0 (>=2.36), libglib2.0-0 (>=2.56), libgpg-error0 (>=1.27), libgraphite2-3 (>=1.3), libgtk-3-0 (>=3.22), libharfbuzz0b (>=1.7), libindicator3-7 (>=16.10), liblz4-1, liblzma5 (>=5.2), libmount1 (>=2.31), libnotify4 (>=0.7), libpango-1.0-0 (>=1.40), libpangocairo-1.0-0 (>=1.40), libpangoft2-1.0-0 (>=1.40), libpcre3 (>=2:8.39), libpixman-1-0 (>=0.34), libpng16-16 (>=1.6), libselinux1 (>=2.7), libsystemd0 (>=237), libthai0 (>=0.1), libuuid1 (>=2.31), libwayland-client0 (>=1.16), libwayland-cursor0 (>=1.16), libwayland-egl1 (>=1.16), libx11-6 (>=2:1.6), libxau6 (>=1:1.0), libxcb1 (>=1.13), libxcb-render0 (>=1.13), libxcb-shm0 (>=1.13), libxcomposite1 (>=1:0.4), libxcursor1 (>=1:1.1), libxdamage1 (>=1:1.1), libxdmcp6 (>=1:1.1), libxext6 (>=2:1.3), libxfixes3 (>=1:5.0), libxi6 (>=2:1.7), libxinerama1 (>=2:1.1), libxkbcommon0 (>=0.8), libxrandr2 (>=2:1.5), libxrender1 (>=1:0.9), zlib1g (>=1:1.2)
Maintainer: Sebastian Bogan <sebogh@qibli.net>
Description: Notify about new Yammer messages.
 Poll the Yammer API and notify about new messages using libnotify.

endef
export DEB_CONTROL

all: goyammer

icons: ./icon/*.png
	cd ./icon; ./build.sh

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
	@echo "$$DEB_CONTROL" > $(DEB_DIR)/DEBIAN/control
	cp goyammer $(DEB_DIR)/usr/bin
	strip --strip-unneeded --remove-section=.comment --remove-section=.note $(DEB_DIR)/usr/bin/goyammer
	cp copyright $(DEB_DIR)/usr/share/doc/goyammer
	cp changelog $(DEB_DIR)/usr/share/doc/goyammer
	gzip --best --no-name $(DEB_DIR)/usr/share/doc/goyammer/changelog
	pandoc goyammer.1.md -s -t man -o $(DEB_DIR)/usr/share/man/man1/goyammer.1
	gzip --best --no-name $(DEB_DIR)/usr/share/man/man1/goyammer.1
	pandoc goyammer-login.1.md -s -t man -o $(DEB_DIR)/usr/share/man/man1/goyammer-login.1
	gzip --best --no-name $(DEB_DIR)/usr/share/man/man1/goyammer-login.1
	pandoc goyammer-poll.1.md -s -t man -o $(DEB_DIR)/usr/share/man/man1/goyammer-poll.1
	gzip --best --no-name $(DEB_DIR)/usr/share/man/man1/goyammer-poll.1


$(DEB_PACKAGE): $(DEB_DIR)
	fakeroot dpkg-deb --build $(DEB_DIR)
	lintian $@ || true

debbuild: tidy $(DEB_PACKAGE)


# use this to find out .deb dependencies
showdebs: goyammer
	./deps.sh

clean:
	go clean github.com/seboghpub/goyammer
	rm -f *~
	rm -Rf $(DEB_DIR_BASE)

tidy: clean
	rm -f goyammer
	rm -f $(DEB_PACKAGE)


.PHONY: all clean tidy debbuild showdebs icons
