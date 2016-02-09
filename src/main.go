package main

import (
	"flag"
	"fmt"
)

const (
	lvmPluginSocketPath = "/var/run/docker/plugins"
)

var (
	flDebug  = flag.Bool("debug", false, "enable debug logging")
	flListen = flag.String("listen", lvmPluginSocketPath+"/lvm.sock", "socket to listen for incoming connections")
)

func main() {
	fmt.Println("Hello Docker LVM Plugin")
}
