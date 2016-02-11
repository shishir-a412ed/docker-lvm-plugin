package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
)

type lvmDriver struct {
	home     string
	vgConfig string
	volumes  map[string]*vol
	count    map[*vol]int
	sync.Mutex
}

type vol struct {
	name       string
	mountPoint string
}

func newDriver(home, vgConfig string) *lvmDriver {
	return &lvmDriver{
		home:     home,
		vgConfig: vgConfig,
		volumes:  make(map[string]*vol),
		count:    make(map[*vol]int),
	}
}

func (l *lvmDriver) Create(req volume.Request) volume.Response {
	l.Lock()
	defer l.Unlock()
	var size string

	if v, exists := l.volumes[req.Name]; exists {
		return resp(v.mountPoint)
	}

	vgName, err := ioutil.ReadFile(l.vgConfig)
	if err != nil {
		return resp(err)
	}

	if len(vgName) == 0 {
		return volume.Response{Err: fmt.Sprintf("Volume group name must be provided for volume creation. Please update the config file %s with volume group name.", l.vgConfig)}
	}

	for key, value := range req.Options {
		if key == "size" {
			size = value
			break
		}
	}

	cmd := exec.Command("lvcreate", "-n", req.Name, "--size", size, strings.Trim(string(vgName), "\n"))
	if err := cmd.Run(); err != nil {
		return resp(err)
	}

	return resp("/home/smahajan/lvm-plugin")
}

func (l *lvmDriver) List(req volume.Request) volume.Response {
	return resp(nil)
}

func (l *lvmDriver) Get(req volume.Request) volume.Response {
	return resp(nil)
}

func (l *lvmDriver) Remove(req volume.Request) volume.Response {
	return resp(nil)
}

func (l *lvmDriver) Path(req volume.Request) volume.Response {
	return resp(nil)
}

func (l *lvmDriver) Mount(req volume.Request) volume.Response {
	fmt.Println("HELLO LVM PLUGIN: MOUNT")
	return resp(nil)
}

func (l *lvmDriver) Unmount(req volume.Request) volume.Response {
	return resp(nil)
}

func resp(r interface{}) volume.Response {
	switch t := r.(type) {
	case error:
		return volume.Response{Err: t.Error()}
	case string:
		return volume.Response{Mountpoint: t}
	default:
		return volume.Response{Err: "bad value writing response"}
	}
}
