package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/juliengk/go-utils/filedir"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type gitVolume struct {
	URL string
	Ref string

	Mountpoint  string
	connections int
}

func (d *volumeDriver) Create(req volume.Request) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Create: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	_, ok := req.Options["url"]
	if !ok {
		res.Err = fmt.Sprintf("url option is mandatory")
		return res
	}
	if len(req.Options["url"]) == 0 {
		res.Err = fmt.Sprintf("url cannot be empty")
		return res
	}

	vol := gitVolume{
		URL:        req.Options["url"],
		Mountpoint: d.getPath(req.Name),
	}

	_, ok = req.Options["ref"]
	if ok {
		vol.Ref = req.Options["ref"]
	}

	if err := d.addVolume(req.Name, &vol); err != nil {
		res.Err = err.Error()
		return res
	}

	d.saveState()

	return res
}

func (d *volumeDriver) List(req volume.Request) volume.Response {
	var res volume.Response

	log.Info("VolumeDriver.List: volumes")

	d.Lock()
	defer d.Unlock()

	res.Volumes = d.listVolumes()

	return res
}

func (d *volumeDriver) Get(req volume.Request) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Get: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	v, err := d.getVolume(req.Name)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	res.Volume = &volume.Volume{
		Name:       req.Name,
		Mountpoint: v.Mountpoint,
	}

	return res
}

func (d *volumeDriver) Remove(req volume.Request) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Remove: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	if err := d.removeVolume(req.Name); err != nil {
		res.Err = err.Error()
		return res
	}

	d.saveState()

	return res
}

func (d *volumeDriver) Path(req volume.Request) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Path: volume %s", req.Name)

	d.RLock()
	defer d.RUnlock()

	_, err := d.getVolume(req.Name)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	res.Mountpoint = d.getPath(req.Name)

	return res
}

func (d *volumeDriver) Mount(req volume.MountRequest) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Mount: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	v, err := d.getVolume(req.Name)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	if v.connections == 0 {
		if err := filedir.CreateDirIfNotExist(v.Mountpoint, true, 0700); err != nil {
			res.Err = err.Error()
			return res
		}

		cloneOpts := &git.CloneOptions{
			URL: v.URL,
		}

		if err := cloneOpts.Validate(); err != nil {
			res.Err = err.Error()
			return res
		}

		r, err := git.PlainClone(v.Mountpoint, false, cloneOpts)
		if err != nil {
			res.Err = err.Error()
			return res
		}

		if len(v.Ref) > 0 {
			w, _ := r.Worktree()
			hash := plumbing.NewHash(v.Ref)

			w.Checkout(&git.CheckoutOptions{
				Hash: hash,
			})
		}
	}

	v.connections++

	res.Mountpoint = v.Mountpoint

	return res
}

func (d *volumeDriver) Unmount(req volume.UnmountRequest) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Unmount: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	v, err := d.getVolume(req.Name)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	v.connections--

	if v.connections <= 0 {
		if err := os.RemoveAll(v.Mountpoint); err != nil {
			res.Err = err.Error()
			return res
		}

		v.connections = 0
	}

	return res
}

func (d *volumeDriver) Capabilities(req volume.Request) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Capabilities: volume %s", req.Name)

	res.Capabilities = volume.Capability{Scope: "local"}

	return res
}
