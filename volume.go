package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/juliengk/go-utils/filedir"
	"github.com/juliengk/go-utils/json"
)

type volumeDriver struct {
	sync.RWMutex

	volPath   string
	statePath string
	volumes   map[string]*gitVolume
}

func NewHandlerFromVolumeDriver(root string) *volume.Handler {
	d := &volumeDriver{
		volPath:   path.Join(root, "volumes"),
		statePath: path.Join(root, "state", "gitfs-state.json"),
		volumes:   map[string]*gitVolume{},
	}

	d.loadState()

	return volume.NewHandler(d)
}

func (d *volumeDriver) loadState() error {
	if filedir.FileExists(d.statePath) {
		data, err := ioutil.ReadFile(d.statePath)
		if err != nil {
			return err
		}

		if err := json.Decode(data, &d.volumes); err != nil {
			return err
		}
	}

	return nil
}

func (d *volumeDriver) saveState() error {
	data := json.Encode(d.volumes)

	if err := ioutil.WriteFile(d.statePath, data.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func (d *volumeDriver) addVolume(name string, vol *gitVolume) error {
	_, ok := d.volumes[name]
	if ok {
		return fmt.Errorf("Volume %s already exists", name)
	}

	d.volumes[name] = vol

	return nil
}

func (d *volumeDriver) removeVolume(name string) error {
	v, err := d.getVolume(name)
	if err != nil {
		return err
	}

	if v.connections > 0 {
		return fmt.Errorf("volume %s is currently used by a container", name)
	}

	if err := os.RemoveAll(v.Mountpoint); err != nil {
		return err
	}

	delete(d.volumes, name)

	return nil
}

func (d *volumeDriver) listVolumes() []*volume.Volume {
	var volumes []*volume.Volume

	for name, v := range d.volumes {
		vol := &volume.Volume{
			Name:       name,
			Mountpoint: v.Mountpoint,
		}

		volumes = append(volumes, vol)
	}

	return volumes
}

func (d *volumeDriver) getVolume(name string) (*gitVolume, error) {
	v, ok := d.volumes[name]
	if !ok {
		return &gitVolume{}, fmt.Errorf("volume %s not found", name)
	}

	return v, nil
}

func (d *volumeDriver) getPath(name string) string {
	return path.Join(d.volPath, name)
}
