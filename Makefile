# Installation Directories
SYSCONFDIR ?=$(DESTDIR)/etc/docker
SYSTEMDIR ?=$(DESTDIR)/usr/lib/systemd/system
GOLANG ?= /usr/bin/go
BINARY ?= docker-lvm-plugin
BINARYLOC ?=$(DESTDIR)/usr/bin

export GO15VENDOREXPERIMENT=1

all: lvm-plugin-build

lvm-plugin-build: main.go driver.go
	$(GOLANG) build -o $(BINARY) .

.PHONY: install
install:
	install docker-lvm-plugin.conf $(SYSCONFDIR)
	install systemd/docker-lvm-plugin.service $(SYSTEMDIR)
	install systemd/docker-lvm-plugin.socket $(SYSTEMDIR)
	install $(BINARY) $(BINARYLOC)

.PHONY: clean
clean:
	rm -rf _vendor
	rm -f $(SYSCONFDIR)/docker-lvm-plugin.conf
	rm -f $(SYSTEMDIR)/docker-lvm-plugin.service
	rm -f $(SYSTEMDIR)/docker-lvm-plugin.socket
	rm -f $(BINARY)
	rm -f $(BINARYLOC)/$(BINARY)

