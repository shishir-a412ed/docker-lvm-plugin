package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
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
	Name       string `json:"name"`
	MountPoint string `json:"mountpoint"`
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
	l.count[v] = 0
	if err := saveToDisk(l.volumes, l.count); err != nil {
		return resp(err)
	}
	return resp(v.MountPoint)
}

func (l *lvmDriver) List(req volume.Request) volume.Response {
	fmt.Println("HELLO LVM PLUGIN: LIST")
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
	fmt.Println("HELLO LVM PLUGIN: GET")
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
	fmt.Println("HELLO LVM PLUGIN: REMOVE")
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

	v := l.volumes[req.Name]
	delete(l.count, v)
	delete(l.volumes, req.Name)
	if err := saveToDisk(l.volumes, l.count); err != nil {
		return resp(err)
	}
	return resp(getMountpoint(l.home, req.Name))
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

	vgName, err := getVolumegroupName(l.vgConfig)
	if err != nil {
		return resp(err)
	}

	cmd := exec.Command("mount", fmt.Sprintf("/dev/%s/%s", vgName, req.Name), getMountpoint(l.home, req.Name))
	if out, err := cmd.CombinedOutput(); err != nil {
		return resp(fmt.Errorf("%s", string(out)))
	}
	return resp(getMountpoint(l.home, req.Name))
}

func (l *lvmDriver) Unmount(req volume.Request) volume.Response {
	fmt.Println("HELLO LVM PLUGIN: UNMOUNT")
	l.Lock()
	defer l.Unlock()
	v := l.volumes[req.Name]
	l.count[v]--
	cmd := exec.Command("umount", getMountpoint(l.home, req.Name))
	if out, err := cmd.CombinedOutput(); err != nil {
		return resp(fmt.Errorf("%s", string(out)))
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
		vgName = str
		break
	}

	return strings.TrimSpace(vgName), nil
}

func getMountpoint(home, name string) string {
	return path.Join(home, name)
}

func saveToDisk(volumes map[string]*vol, count map[*vol]int) error {
	fhVolumes, err := os.Create(lvmVolumesConfigPath)
	if err != nil {
		return err
	}
	defer fhVolumes.Close()

	if err := json.NewEncoder(fhVolumes).Encode(&volumes); err != nil {
		return err
	}

	fhCount, err := os.Create(lvmCountConfigPath)
	if err != nil {
		return err
	}
	defer fhCount.Close()

	csvWriter := csv.NewWriter(fhCount)
	for v, c := range count {
		if err := csvWriter.Write([]string{v.Name, v.MountPoint, strconv.Itoa(c)}); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return nil
}

func loadFromDisk(l *lvmDriver) error {
	jsonVolumes, err := os.Open(lvmVolumesConfigPath)
	if err != nil {
		return err
	}
	defer jsonVolumes.Close()

	// Load volume store metadata
	if err := json.NewDecoder(jsonVolumes).Decode(&l.volumes); err != nil {
		return err
	}

	// Load count store metadata
	fhRead, err := os.Open(lvmCountConfigPath)
	if err != nil {
		return err
	}

	defer fhRead.Close()

	csvReader := csv.NewReader(fhRead)
	for {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		v := &vol{Name: record[0], MountPoint: record[1]}
		c, err := strconv.Atoi(record[2])
		if err != nil {
			return err
		}
		l.count[v] = c
	}

	return nil
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
