package main

import (
	log "github.com/Sirupsen/logrus"
)

func main() {
	h := NewHandlerFromVolumeDriver("/var/lib/docker")
	log.Fatal(h.ServeUnix("gitvol", 1))
}
