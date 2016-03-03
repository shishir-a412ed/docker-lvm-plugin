package main

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	count    map[string]int
	sync.RWMutex
}

type vol struct {
	Name       string `json:"name"`
	MountPoint string `json:"mountpoint"`
}

func newDriver(home, vgConfig string) *lvmDriver {
	return &lvmDriver{
		home:     home,
		vgConfig: vgConfig,
		volumes:  make(map[string]*vol),
		count:    make(map[string]int),
	}
}

func (l *lvmDriver) Create(req volume.Request) volume.Response {
	l.Lock()
	defer l.Unlock()
	var size string

	if v, exists := l.volumes[req.Name]; exists {
		return resp(v.MountPoint)
	}

	vgName, err := getVolumegroupName(l.vgConfig)
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

	cmd := exec.Command("lvcreate", "-n", req.Name, "--size", size, vgName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return resp(fmt.Errorf("%s", string(out)))
	}

	cmd = exec.Command("mkfs.xfs", fmt.Sprintf("/dev/%s/%s", vgName, req.Name))
	if out, err := cmd.CombinedOutput(); err != nil {
		return resp(fmt.Errorf("%s", string(out)))
	}

	mp := getMountpoint(l.home, req.Name)
	if err := os.MkdirAll(mp, 0700); err != nil {
		return resp(err)
	}

	v := &vol{req.Name, mp}
	l.volumes[v.Name] = v
	l.count[v.Name] = 0
	if err := saveToDisk(l.volumes, l.count); err != nil {
		return resp(err)
	}
	return resp(v.MountPoint)
}

func (l *lvmDriver) List(req volume.Request) volume.Response {
	var res volume.Response
	l.Lock()
	defer l.Unlock()
	var ls []*volume.Volume
	for _, vol := range l.volumes {
		v := &volume.Volume{
			Name:       vol.Name,
			Mountpoint: vol.MountPoint,
		}
		ls = append(ls, v)
	}
	res.Volumes = ls
	return res
}

func (l *lvmDriver) Get(req volume.Request) volume.Response {
	var res volume.Response
	l.Lock()
	defer l.Unlock()
	v, exists := l.volumes[req.Name]
	if !exists {
		return resp(fmt.Errorf("no such volume"))
	}
	res.Volume = &volume.Volume{
		Name:       v.Name,
		Mountpoint: v.MountPoint,
	}
	return res
}

func (l *lvmDriver) Remove(req volume.Request) volume.Response {
	l.Lock()
	defer l.Unlock()

	if err := os.RemoveAll(getMountpoint(l.home, req.Name)); err != nil {
		return resp(err)
	}

	vgName, err := getVolumegroupName(l.vgConfig)
	if err != nil {
		return resp(err)
	}

	cmd := exec.Command("lvremove", "--force", fmt.Sprintf("%s/%s", vgName, req.Name))
	if out, err := cmd.CombinedOutput(); err != nil {
		return resp(fmt.Errorf("%s", string(out)))
	}

	delete(l.count, req.Name)
	delete(l.volumes, req.Name)
	if err := saveToDisk(l.volumes, l.count); err != nil {
		return resp(err)
	}
	return resp(getMountpoint(l.home, req.Name))
}

func (l *lvmDriver) Path(req volume.Request) volume.Response {
	return resp(getMountpoint(l.home, req.Name))
}

func (l *lvmDriver) Mount(req volume.Request) volume.Response {
	l.Lock()
	defer l.Unlock()
	l.count[req.Name]++
	vgName, err := getVolumegroupName(l.vgConfig)
	if err != nil {
		return resp(err)
	}
	if l.count[req.Name] == 1 {
		cmd := exec.Command("mount", fmt.Sprintf("/dev/%s/%s", vgName, req.Name), getMountpoint(l.home, req.Name))
		if out, err := cmd.CombinedOutput(); err != nil {
			return resp(fmt.Errorf("%s", string(out)))
		}
	}
	if err := saveToDisk(l.volumes, l.count); err != nil {
		return resp(err)
	}
	return resp(getMountpoint(l.home, req.Name))
}

func (l *lvmDriver) Unmount(req volume.Request) volume.Response {
	l.Lock()
	defer l.Unlock()
	l.count[req.Name]--
	if l.count[req.Name] == 0 {
		cmd := exec.Command("umount", getMountpoint(l.home, req.Name))
		if out, err := cmd.CombinedOutput(); err != nil {
			return resp(fmt.Errorf("%s", string(out)))
		}
	}
	if err := saveToDisk(l.volumes, l.count); err != nil {
		return resp(err)
	}
	return resp(getMountpoint(l.home, req.Name))
}

func getVolumegroupName(vgConfig string) (string, error) {
	vgName := ""
	inFile, err := os.Open(vgConfig)
	if err != nil {
		return "", err
	}
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		str := scanner.Text()
		if strings.HasPrefix(str, "#") {
			continue
		}
		vgSlice := strings.Split(str, "=")
		vgName = vgSlice[1]
		break
	}

	return strings.TrimSpace(vgName), nil
}

func getMountpoint(home, name string) string {
	return path.Join(home, name)
}

func saveToDisk(volumes map[string]*vol, count map[string]int) error {
	// Save volume store metadata.
	fhVolumes, err := os.Create(lvmVolumesConfigPath)
	if err != nil {
		return err
	}
	defer fhVolumes.Close()

	if err := json.NewEncoder(fhVolumes).Encode(&volumes); err != nil {
		return err
	}

	// Save count store metadata.
	fhCount, err := os.Create(lvmCountConfigPath)
	if err != nil {
		return err
	}
	defer fhCount.Close()

	return json.NewEncoder(fhCount).Encode(&count)
}

func loadFromDisk(l *lvmDriver) error {
	// Load volume store metadata
	jsonVolumes, err := os.Open(lvmVolumesConfigPath)
	if err != nil {
		return err
	}
	defer jsonVolumes.Close()

	if err := json.NewDecoder(jsonVolumes).Decode(&l.volumes); err != nil {
		return err
	}

	// Load count store metadata
	jsonCount, err := os.Open(lvmCountConfigPath)
	if err != nil {
		return err
	}
	defer jsonCount.Close()

	return json.NewDecoder(jsonCount).Decode(&l.count)
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
