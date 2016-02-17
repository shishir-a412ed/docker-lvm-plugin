package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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

	fmt.Println("HELLO LVM PLUGIN: CREATE")

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

	mp := getMountpoint(l.home, req.Name)
	if err := os.MkdirAll(mp, 0700); err != nil {
		return resp(err)
	}

	v := &vol{req.Name, mp}
	l.volumes[v.name] = v
	l.count[v] = 0
	return resp(v.mountPoint)
}

func (l *lvmDriver) List(req volume.Request) volume.Response {
	fmt.Println("HELLO LVM PLUGIN: LIST")
	var res volume.Response
	l.Lock()
	defer l.Unlock()
	var ls []*volume.Volume
	fmt.Println(len(ls))
	for _, vol := range l.volumes {
		v := &volume.Volume{
			Name:       vol.name,
			Mountpoint: vol.mountPoint,
		}
		ls = append(ls, v)
	}
	res.Volumes = ls
	return res
}

func (l *lvmDriver) Get(req volume.Request) volume.Response {
	fmt.Println("HELLO LVM PLUGIN: GET")
	var res volume.Response
	l.Lock()
	defer l.Unlock()
	v, exists := l.volumes[req.Name]
	if !exists {
		return resp(fmt.Errorf("no such volume"))
	}
	res.Volume = &volume.Volume{
		Name:       v.name,
		Mountpoint: v.mountPoint,
	}
	return res
}

func (l *lvmDriver) Remove(req volume.Request) volume.Response {
	return resp(nil)
}

func (l *lvmDriver) Path(req volume.Request) volume.Response {
	fmt.Println("HELLO LVM PLUGIN: PATH")
	return resp(getMountpoint(l.home, req.Name))
}

func (l *lvmDriver) Mount(req volume.Request) volume.Response {
	fmt.Println("HELLO LVM PLUGIN: MOUNT")
	l.Lock()
	defer l.Unlock()
	v := l.volumes[req.Name]
	l.count[v]++
	return resp(getMountpoint(l.home, req.Name))
}

func (l *lvmDriver) Unmount(req volume.Request) volume.Response {
	fmt.Println("HELLO LVM PLUGIN: UNMOUNT")
	l.Lock()
	defer l.Unlock()
	v := l.volumes[req.Name]
	l.count[v]--
	return resp(getMountpoint(l.home, req.Name))
}

func getMountpoint(home, name string) string {
	return path.Join(home, name)
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
