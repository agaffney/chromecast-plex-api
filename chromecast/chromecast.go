package chromecast

import (
	"context"
	"github.com/agaffney/chromecast-plex-api/config"
	castdns "github.com/vishen/go-chromecast/dns"
	"log"
	"net"
	"time"
)

type DeviceEntry struct {
	Device     string `json:"device"`
	DeviceName string `json:"name"`
	Address    net.IP `json:"address"`
	Port       int    `json:"port"`
	UUID       string `json:"uuid"`
}

var devices []*DeviceEntry

func StartScanning() {
	go func() {
		config := config.Get()
		devices = make([]*DeviceEntry, 0)
		for {
			var iface *net.Interface
			var err error
			if config.CastInterface != "" {
				if iface, err = net.InterfaceByName(config.CastInterface); err != nil {
					log.Fatalf("unable to find interface %q: %v", config.CastInterface, err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(config.CastDnsTimeout))
			defer cancel()
			castEntryChan, err := castdns.DiscoverCastDNSEntries(ctx, iface)
			if err != nil {
				log.Fatalf("Error while discovering cast devices: %s", err)
			}
			for d := range castEntryChan {
				foundDevice := false
				for _, tmpDevice := range devices {
					if tmpDevice.UUID == d.UUID {
						foundDevice = true
						break
					}
				}
				if !foundDevice {
					deviceEntry := &DeviceEntry{
						Device:     d.Device,
						DeviceName: d.DeviceName,
						Address:    d.AddrV4,
						Port:       d.Port,
						UUID:       d.UUID,
					}
					devices = append(devices, deviceEntry)
					log.Printf("Discovered cast device: %#v", deviceEntry)
				}
			}
			time.Sleep(time.Second * time.Duration(config.CastScanInterval))
		}
	}()
}

func GetDevices() []*DeviceEntry {
	return devices
}
