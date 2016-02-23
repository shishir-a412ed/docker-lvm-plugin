# Installation Directories
SYSCONFDIR ?= /etc/sysconfig
SYSTEMDIR ?= /usr/lib/systemd/system
GOLANG ?= /usr/bin/go
BINARY ?= lvm-plugin
GOPATH ?= $(CURDIR)/_vendor
export GOPATH

all: lvm-plugin-build

lvm-plugin-build: src/main.go src/driver.go
	$(GOLANG) get github.com/Sirupsen/logrus
	$(GOLANG) get github.com/docker/go-plugins-helpers/volume
	$(GOLANG) build -o $(BINARY) src/main.go src/driver.go
	

.PHONY: install 
install: all
	cp docker-lvm-volumegroup $(SYSCONFDIR)
	cp lvm-plugin.service $(SYSTEMDIR)
	mv $(BINARY) /usr/bin

.PHONY: clean
clean:
	rm -rf _vendor
	rm $(SYSCONFDIR)/docker-lvm-volumegroup
	rm $(SYSTEMDIR)/lvm-plugin.service
	rm /usr/bin/$(BINARY)

