package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/juliengk/go-git"
	"github.com/juliengk/go-utils"
	"github.com/juliengk/go-utils/filedir"
	"github.com/kassisol/libsecret"
	log "github.com/sirupsen/logrus"
)

type gitVolume struct {
	URL  string
	Ref  string
	Auth auth

	Mountpoint  string
	connections int
}

type auth struct {
	Type   string
	Driver string
	User   string
	Config map[string]string
}

func (d *volumeDriver) Create(req *volume.CreateRequest) error {
	var secretDriver string

	allowedAuthTypes := []string{
		"anonymous",
		"password",
		"pubkey",
		"token",
	}

	optsSkip := []string{
		"url",
		"ref",
		"auth-type",
		"auth-user",
		"secret-driver",
	}

	supportedTransportSchemes := []string{
		"http",
		"https",
		"ssh",
	}

	log.Infof("VolumeDriver.Create: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	if _, ok := req.Options["url"]; !ok {
		return fmt.Errorf("url option is mandatory")
	}
	if len(req.Options["url"]) == 0 {
		return fmt.Errorf("url cannot be empty")
	}

	u, err := url.Parse(req.Options["url"])
	if err != nil {
		return err
	}

	if !utils.StringInSlice(u.Scheme, supportedTransportSchemes, false) {
		return fmt.Errorf("url transport scheme is not valid. Valid types are %s.", strings.Join(supportedTransportSchemes, ", "))
	}

	authType := "anonymous"
	if v, ok := req.Options["auth-type"]; ok {
		authType = v
	}

	if !utils.StringInSlice(authType, allowedAuthTypes, false) {
		return fmt.Errorf("auth-type is not valid. Valid types are %s.", strings.Join(allowedAuthTypes, ", "))
	}

	if authType != "anonymous" {
		if _, ok := req.Options["auth-user"]; !ok {
			return fmt.Errorf("auth-user option should be set")
		}

		secretDriver = "stdin"
		if v, ok := req.Options["secret-driver"]; ok {
			secretDriver = v
		}
	}

	vol := gitVolume{
		URL:        req.Options["url"],
		Auth:       auth{Type: authType, Driver: secretDriver},
		Mountpoint: d.getPath(req.Name),
	}

	if _, ok := req.Options["ref"]; ok {
		vol.Ref = req.Options["ref"]
	}

	if authType != "anonymous" {
		vol.Auth.User = req.Options["auth-user"]

		configs := make(map[string]string)

		sec, err := libsecret.NewDriver(secretDriver)
		if err != nil {
			return err
		}

		for k, v := range req.Options {
			if !utils.StringInSlice(k, optsSkip, false) {
				sec.AddKey(k, v)

				configs[k] = v
			}
		}

		if err := sec.ValidateKeys(); err != nil {
			return err
		}

		vol.Auth.Config = configs
	}

	if err := d.addVolume(req.Name, &vol); err != nil {
		return err
	}

	d.saveState()

	return nil
}

func (d *volumeDriver) List() (*volume.ListResponse, error) {
	log.Info("VolumeDriver.List: volumes")

	d.Lock()
	defer d.Unlock()

	return &volume.ListResponse{
		Volumes: d.listVolumes(),
	}, nil
}

func (d *volumeDriver) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	log.Infof("VolumeDriver.Get: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	v, err := d.getVolume(req.Name)
	if err != nil {
		return nil, err
	}

	return &volume.GetResponse{
		Volume: &volume.Volume{
			Name:       req.Name,
			Mountpoint: v.Mountpoint,
		},
	}, nil
}

func (d *volumeDriver) Remove(req *volume.RemoveRequest) error {
	log.Infof("VolumeDriver.Remove: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	if err := d.removeVolume(req.Name); err != nil {
		return err
	}

	d.saveState()

	return nil
}

func (d *volumeDriver) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	log.Infof("VolumeDriver.Path: volume %s", req.Name)

	d.RLock()
	defer d.RUnlock()

	if _, err := d.getVolume(req.Name);  err != nil {
		return nil, err
	}

	return &volume.PathResponse{
		Mountpoint: d.getPath(req.Name),
	}, nil
}

func (d *volumeDriver) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	log.Infof("VolumeDriver.Mount: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	v, err := d.getVolume(req.Name)
	if err != nil {
		return nil, err
	}

	if v.connections == 0 {
		g, err := git.New(v.URL)
		if err != nil {
			return nil, err
		}

		if v.Auth.Type != "anonymous" {
			sec, err := libsecret.NewDriver(v.Auth.Driver)
			if err != nil {
				return nil, err
			}

			for k, v := range v.Auth.Config {
				sec.AddKey(k, v)
			}

			secr8, err := sec.GetSecret()
			if err != nil {
				return nil, err
			}

			if err := g.SetAuth(v.Auth.User, v.Auth.Type, secr8); err != nil {
				return nil, err
			}
		}

		if err := filedir.CreateDirIfNotExist(v.Mountpoint, true, 0755); err != nil {
			return nil, err
		}

		if err = g.Clone(v.Mountpoint); err != nil {
			return nil, err
		}

		if len(v.Ref) > 0 {
			if err := g.Checkout(v.Ref); err != nil {
				return nil, err
			}
		}
	}

	v.connections++

	return &volume.MountResponse{
		Mountpoint: v.Mountpoint,
	}, nil
}

func (d *volumeDriver) Unmount(req *volume.UnmountRequest) error {
	log.Infof("VolumeDriver.Unmount: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	v, err := d.getVolume(req.Name)
	if err != nil {
		return err
	}

	v.connections--

	if v.connections <= 0 {
		if err := os.RemoveAll(v.Mountpoint); err != nil {
			return err
		}

		v.connections = 0
	}

	return nil
}

func (d *volumeDriver) Capabilities() *volume.CapabilitiesResponse {
	log.Infof("VolumeDriver.Capabilities: volume")

	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{Scope: "local"},
	}
}
