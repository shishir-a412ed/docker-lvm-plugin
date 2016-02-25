# Installation Directories
SYSCONFDIR ?= /etc/docker
SYSTEMDIR ?= /usr/lib/systemd/system
GOLANG ?= /usr/bin/go
BINARY ?= docker-lvm-plugin
GOPATH ?= $(CURDIR)/_vendor
export GOPATH

all: lvm-plugin-build

lvm-plugin-build: src/main.go src/driver.go
	$(GOLANG) get github.com/Sirupsen/logrus
	$(GOLANG) get github.com/docker/go-plugins-helpers/volume
	$(GOLANG) build -o $(BINARY) src/main.go src/driver.go
	

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

