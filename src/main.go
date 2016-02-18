package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
)

const (
	lvmPluginSocketPath = "/var/run/docker/plugins"
	vgConfigPath        = "/etc/sysconfig/docker-lvm-volumegroup"
	lvmConfigPath       = "/etc/lvmConfig.json"
)

var (
	flVgPath  *string
	flHome    *string
	flVersion *bool
	flDebug   *bool
	flListen  *string
)

func init() {
	flVgPath = flag.String("vg-config", vgConfigPath, "Location of the volume group config file")
	flHome = flag.String("home", "/var/run/docker-lvm", "Home directory for lvm volume storage")
	flVersion = flag.Bool("version", false, "Print version information and quit")
	flDebug = flag.Bool("debug", false, "Enable debug logging")
	flListen = flag.String("listen", lvmPluginSocketPath+"/lvm.sock", "Socket to listen for incoming connections")
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

	if _, err := os.Stat(*flHome); err != nil {
		if !os.IsNotExist(err) {
			logrus.Fatal(err)
		}
		logrus.Debugf("Created home dir at %s", *flHome)
		if err := os.MkdirAll(*flHome, 0700); err != nil {
			logrus.Fatal(err)
		}
	}

	if _, err := os.Stat(*flVgPath); os.IsNotExist(err) {
		file, err := os.Create(*flVgPath)
		if err != nil {
			logrus.Fatal(err)
		}
		if err := file.Chmod(0700); err != nil {
			logrus.Fatal(err)
		}
	}

	lvm := newDriver(*flHome, *flVgPath)
	if err := loadFromDisk(lvm); err != nil {
		logrus.Fatal(err)
	}

	h := volume.NewHandler(lvm)
	if err := h.ServeUnix("root", *flListen); err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("docker lvm plugin successful")
}
