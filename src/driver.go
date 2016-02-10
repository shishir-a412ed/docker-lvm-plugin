package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type lvmDriver struct {
	home     string
	vgConfig string
	volumes  map[string]*vol
	count    map[*vol]int
	sync.Mutex
}

type volume struct {
	name       string
	mountPoint string
}

func newDriver(home, vgConfig string) *lvmDriver {
	return &lvmDriver{
		home:     home,
		vgConfig: vgConfig,
	}
}
