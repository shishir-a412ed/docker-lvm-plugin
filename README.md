# docker lvm driver
Docker Volume Driver for lvm volumes

This plugin can be used to create lvm volumes of specified size, which can 
then be bind mounted into the container using `docker run` command.

## Setup

	1) git clone git@github.com:shishir-a412ed/docker-lvm-driver.git
	2) cd docker-lvm-driver
	3) sudo make install

## Usage

1) Start the docker daemon before starting the lvm-plugin daemon.
   You can start docker daemon using `systemctl start docker`

2) Once docker daemon is up and running, you can start lvm-plugin daemon
   using `systemctl start lvm-plugin`

3) Since logical volumes (lv's) are based on a volume group, it is the 
   responsibility of the user (administrator) to provide a volume group name.
   You can choose an existing volume group name by listing volume groups on 
   your system using `vgs` command OR create a new volume group using 
   `vgcreate` command.

4) Update volume group name in the config file `/etc/sysconfig/docker-lvm-volumegroup`

## Volume Creation

``` bash
$ docker volume create -d lvm --name foobar --opt size=0.2G
```
This will create a lvm volume named foobar of size 208 MB (0.2 GB).

## Volume List

``` bash
$ docker volume ls
```
This will list volumes created by all docker drivers including the default driver (local).

## Volume Inspect

``` bash
$ docker volume inspect foobar
```
This will inspect foobar and return a JSON.
```bash
[
    {
        "Name": "foobar",
        "Driver": "lvm",
        "Mountpoint": "/run/docker-lvm/foobar123"
    }
]
```

## Volume Removal
```bash
$ docker volume rm foobar
```
This will remove lvm volume foobar.

## Bind Mount lvm volume inside the container

```bash
$ docker run -it --volume-driver=lvm -v foobar:/home fedora /bin/bash
```
This will bind mount the logical volume `foobar` into the home directory of the container.

## License
Apache





