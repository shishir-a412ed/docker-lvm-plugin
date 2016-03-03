# Installation Directories
SYSCONFDIR ?= /etc/docker
SYSTEMDIR ?= /usr/lib/systemd/system
GOLANG ?= /usr/bin/go
BINARY ?= docker-lvm-plugin
export GOPATH := $(CURDIR)/Godeps/_workspace:$(GOPATH)

all: lvm-plugin-build

lvm-plugin-build: main.go driver.go
	$(GOLANG) build -o $(BINARY) main.go driver.go
	

.PHONY: install 
install: all
	cp docker-lvm-plugin.conf $(SYSCONFDIR)
	cp docker-lvm-plugin.service $(SYSTEMDIR)
	mv $(BINARY) /usr/bin

.PHONY: clean
clean:
	rm -rf _vendor
	rm $(SYSCONFDIR)/docker-lvm-plugin.conf
	rm $(SYSTEMDIR)/docker-lvm-plugin.service
	rm /usr/bin/$(BINARY)

