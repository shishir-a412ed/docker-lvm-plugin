# Installation Directories
SYSCONFDIR ?= /etc/docker
SYSTEMDIR ?= /usr/lib/systemd/system
GOLANG ?= /usr/bin/go
BINARY ?= docker-lvm-plugin
export GO15VENDOREXPERIMENT=1

all: lvm-plugin-build

lvm-plugin-build: main.go driver.go
	$(GOLANG) build -o $(BINARY) .
	

.PHONY: install 
install: all
	cp docker-lvm-plugin.conf $(SYSCONFDIR)
	cp docker-lvm-plugin.service $(SYSTEMDIR)
	mv $(BINARY) /usr/bin

.PHONY: clean
clean:
	rm -rf _vendor
	rm -f $(SYSCONFDIR)/docker-lvm-plugin.conf
	rm -f $(SYSTEMDIR)/docker-lvm-plugin.service
	rm -f $(BINARY)
	rm -f /usr/bin/$(BINARY)

