package plex

import (
	"github.com/agaffney/chromecast-plex-api/chromecast/device"
)

const (
	plexAppId     = "9AC194DC"
	plexNamespace = "urn:x-cast:plex"
)

type Controller struct {
	device *device.Device
}

func NewController(device *device.Device) *Controller {
	c := &Controller{device: device}
	return c
}

func (c *Controller) Launch() error {
	return c.device.Launch(plexAppId)
}

func (c *Controller) Next() error {
	// TODO: do something
	return nil
}
