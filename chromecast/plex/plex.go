package plex

import (
	"github.com/agaffney/chromecast-plex-api/chromecast/device"
	"github.com/vishen/go-chromecast/cast"
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
	if err := c.device.Launch(plexAppId); err != nil {
		return err
	}
	_, err := c.device.Send(&cast.GetStatusHeader, "sender-0", c.device.GetApplication().TransportId, plexNamespace)
	return err
}

func (c *Controller) Next() error {
	_, err := c.device.Send(&cast.PayloadHeader{Type: "NEXT"}, "sender-0", c.device.GetApplication().TransportId, plexNamespace)
	return err
}
