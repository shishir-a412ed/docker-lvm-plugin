package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
)

const (
	lvmPluginSocketPath  = "/run/docker/plugins/lvm.sock"
	vgConfigPath         = "/etc/sysconfig/docker-lvm-volumegroup"
	lvmHome              = "/run/docker-lvm"
	lvmVolumesConfigPath = "/etc/lvmVolumesConfig.json"
	lvmCountConfigPath   = "/etc/lvmCountConfig.csv"
)

var (
	flVersion *bool
	flDebug   *bool
)

func init() {
	flVersion = flag.Bool("version", false, "Print version information and quit")
	flDebug = flag.Bool("debug", false, "Enable debug logging")
}

func cleanup() error {
	return os.Remove(lvmPluginSocketPath)
}

func main() {

	flag.Parse()

	if *flVersion {
		fmt.Fprint(os.Stdout, "docker lvm plugin version: 1.0\n")
		return
	}

	if *flDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if _, err := os.Stat(lvmHome); err != nil {
		if !os.IsNotExist(err) {
			logrus.Fatal(err)
		}
		logrus.Debugf("Created home dir at %s", lvmHome)
		if err := os.MkdirAll(lvmHome, 0700); err != nil {
			logrus.Fatal(err)
		}
	}

	lvm := newDriver(lvmHome, vgConfigPath)

	// Call loadFromDisk only if config file exists.
	if _, err := os.Stat(lvmVolumesConfigPath); err == nil {
		if err := loadFromDisk(lvm); err != nil {
			logrus.Fatal(err)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		if err := cleanup(); err != nil {
			logrus.Fatal(err)
		}
		os.Exit(0)
	}()

	h := volume.NewHandler(lvm)
	if err := h.ServeUnix("root", lvmPluginSocketPath); err != nil {
		logrus.Fatal(err)
	}
}
