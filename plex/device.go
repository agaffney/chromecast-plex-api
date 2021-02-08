package plex

import (
	"fmt"
	"github.com/agaffney/chromecast-plex-api/chromecast"
	castdevice "github.com/agaffney/chromecast-plex-api/chromecast/device"
	"github.com/agaffney/chromecast-plex-api/config"
	"github.com/jrudio/go-plex-client"
	"strings"
	"time"
)

const (
	DEVICE_TYPE_PLEX_PLAYER = "plex_player"
	DEVICE_TYPE_CHROMECAST  = "chromecast"
)

type Device struct {
	Type             string     `json:"type"`
	Id               string     `json:"id"`
	Name             string     `json:"name"`
	FirstSeen        *time.Time `json:"first_seen"`
	LastSeen         *time.Time `json:"last_seen"`
	chromecastDevice *castdevice.Device
	plexDevice       *plex.PMSDevices
}

var cfg *config.Config

func (p *Plex) GetDevices() []*Device {
	return p.devices
}

func (p *Plex) GetDevice(id string) *Device {
	for _, device := range p.devices {
		if device.Id == id {
			return device
		}
	}
	return nil
}

func (p *Plex) RefreshDevices() error {
	p.refreshDevicesMutex.Lock()
	defer p.refreshDevicesMutex.Unlock()
	// Plex players
	plexDevices, err := p.conn.GetDevices()
	if err != nil {
		return err
	}
	for _, device := range plexDevices {
		if device.Presence == "0" {
			continue
		}
		timeNow := time.Now()
		foundDevice := false
		for idx, tmpDevice := range p.devices {
			if device.ClientIdentifier == tmpDevice.Id {
				p.devices[idx].LastSeen = &timeNow
				foundDevice = true
				break
			}
		}
		if !foundDevice {
			provides := strings.Split(device.Provides, ",")
			isPlayer := false
			for _, foo := range provides {
				if foo == "player" {
					isPlayer = true
					break
				}
			}
			if !isPlayer {
				continue
			}
			d := &Device{
				Type:       DEVICE_TYPE_PLEX_PLAYER,
				Id:         device.ClientIdentifier,
				Name:       fmt.Sprintf("%s - %s", device.Name, device.Product),
				FirstSeen:  &timeNow,
				LastSeen:   &timeNow,
				plexDevice: &device,
			}
			p.devices = append(p.devices, d)
		}
	}
	// Chromecast devices
	castDevices := chromecast.GetDevices()
	for _, device := range castDevices {
		timeNow := time.Now()
		foundDevice := false
		for idx, tmpDevice := range p.devices {
			if device.UUID == tmpDevice.Id {
				p.devices[idx].LastSeen = &timeNow
				foundDevice = true
				break
			}
		}
		if !foundDevice {
			d := &Device{
				Type:             DEVICE_TYPE_CHROMECAST,
				Id:               device.UUID,
				Name:             device.Name,
				FirstSeen:        &timeNow,
				LastSeen:         &timeNow,
				chromecastDevice: device,
			}
			p.devices = append(p.devices, d)
		}
	}
	return nil
}
