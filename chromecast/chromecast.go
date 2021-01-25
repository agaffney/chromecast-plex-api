package chromecast

import (
	"context"
	"github.com/agaffney/chromecast-plex-api/chromecast/device"
	"github.com/agaffney/chromecast-plex-api/config"
	castdns "github.com/vishen/go-chromecast/dns"
	"log"
	"net"
	"sync"
	"time"
)

var devices []*device.Device
var scanMutex sync.Mutex
var cfg *config.Config

func StartScanning() {
	cfg = config.Get()
	devices = make([]*device.Device, 0)
	go func() {
		for {
			Scan()
			time.Sleep(time.Second * time.Duration(cfg.CastScanInterval))
		}
	}()
}

func Scan() {
	var iface *net.Interface
	var err error
	if cfg.CastInterface != "" {
		if iface, err = net.InterfaceByName(cfg.CastInterface); err != nil {
			log.Fatalf("unable to find interface %q: %v", cfg.CastInterface, err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.CastDnsTimeout))
	defer cancel()
	scanMutex.Lock()
	castEntryChan, err := castdns.DiscoverCastDNSEntries(ctx, iface)
	if err != nil {
		log.Fatalf("Error while discovering cast devices: %s", err)
	}
	for d := range castEntryChan {
		foundDevice := false
		for idx, tmpDevice := range devices {
			if tmpDevice.UUID == d.UUID {
				foundDevice = true
				timeNow := time.Now()
				devices[idx].LastSeen = &timeNow
				break
			}
		}
		if !foundDevice {
			timeNow := time.Now()
			deviceEntry := &device.Device{
				Device:     d.Device,
				DeviceName: d.DeviceName,
				Address:    d.AddrV4,
				Port:       d.Port,
				UUID:       d.UUID,
				FirstSeen:  &timeNow,
				LastSeen:   &timeNow,
			}
			devices = append(devices, deviceEntry)
		}
	}
	scanMutex.Unlock()
}

func GetDevices() []*device.Device {
	return devices
}

func GetDevice(uuid string) *device.Device {
	for _, device := range devices {
		if device.UUID == uuid {
			return device
		}
	}
	return nil
}
