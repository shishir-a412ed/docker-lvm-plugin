package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	lvmPluginSocketPath = "/var/run/docker/plugins"
	vgConfigPath        = "/etc/sysconfig/docker-lvm-driver"
)

var (
	flVgPath  *string
	flVersion *bool
	flListen  *string
)

func init() {
	flVgPath = flag.String("vg-config", vgConfigPath, "Location of the volume group config file")
	flVersion = flag.Bool("version", false, "Print version information and quit")
	flListen = flag.String("listen", lvmPluginSocketPath+"/lvm.sock", "socket to listen for incoming connections")
}

func main() {

	flag.Parse()

	if *flVersion {
		fmt.Fprint(os.Stdout, "docker lvm plugin version: 1.0\n")
		return
	}

	fmt.Println("docker lvm plugin successful")
}
