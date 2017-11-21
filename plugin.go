package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/juliengk/go-utils"
	"github.com/juliengk/go-utils/filedir"
	"github.com/kassisol/docker-volume-git/secret"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
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

func (d *volumeDriver) Create(req volume.Request) volume.Response {
	var res volume.Response
	var secretDriver string

	allowedAuthTypes := []string{
		"anonymous",
		"password",
		"pubkey",
	}

	optsSkip := []string{
		"url",
		"ref",
		"auth-type",
		"auth-user",
		"secret-driver",
	}

	log.Infof("VolumeDriver.Create: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	if _, ok := req.Options["url"]; !ok {
		res.Err = fmt.Sprintf("url option is mandatory")
		return res
	}
	if len(req.Options["url"]) == 0 {
		res.Err = fmt.Sprintf("url cannot be empty")
		return res
	}

	authType := "anonymous"
	if v, ok := req.Options["auth-type"]; ok {
		authType = v
	}

	if !utils.StringInSlice(authType, allowedAuthTypes, false) {
		res.Err = fmt.Sprintf("auth-type is not valid. Valid types are %s.", strings.Join(allowedAuthTypes, ", "))
		return res
	}

	if authType != "anonymous" {
		if _, ok := req.Options["auth-user"]; !ok {
			res.Err = fmt.Sprintf("auth-user option should be set")
			return res
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
	}

	configs := make(map[string]string)

	sec, err := secret.NewDriver(secretDriver)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	for k, v := range req.Options {
		if !utils.StringInSlice(k, optsSkip, false) {
			sec.AddKey(k, v)

			configs[k] = v
		}
	}

	if err := sec.ValidateKeys(); err != nil {
		res.Err = err.Error()
		return res
	}

	vol.Auth.Config = configs

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
	var a gitssh.AuthMethod

	log.Infof("VolumeDriver.Mount: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	v, err := d.getVolume(req.Name)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	if v.connections == 0 {
		cloneOpts := &git.CloneOptions{
			URL: v.URL,
		}

		if v.Auth.Type != "anonymous" {
			sec, err := secret.NewDriver(v.Auth.Driver)
			if err != nil {
				res.Err = err.Error()
				return res
			}

			for k, v := range v.Auth.Config {
				sec.AddKey(k, v)
			}

			secr8, err := sec.GetSecret()
			if err != nil {
				res.Err = err.Error()
				return res
			}

			if v.Auth.Type == "password" {
				a = &gitssh.Password{User: v.Auth.User, Pass: secr8}
			}

			if v.Auth.Type == "pubkey" {
				a, err = gitssh.NewPublicKeys(v.Auth.User, []byte(secr8), "")
				if err != nil {
					res.Err = err.Error()
					return res
				}
			}

			a.(*gitssh.PublicKeys).HostKeyCallback = ssh.InsecureIgnoreHostKey()
			cloneOpts.Auth = a
		}

		if err := filedir.CreateDirIfNotExist(v.Mountpoint, true, 0700); err != nil {
			res.Err = err.Error()
			return res
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
