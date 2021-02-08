package plex

import (
	"github.com/agaffney/chromecast-plex-api/config"
	"github.com/jrudio/go-plex-client"
	"sync"
	"time"
)

type Plex struct {
	conn                *plex.Plex
	devices             []*Device
	refreshDevicesMutex sync.Mutex
}

var plexObj *Plex

func New(baseUrl string, token string) (*Plex, error) {
	plexObj = &Plex{}
	var err error
	plexObj.conn, err = plex.New(baseUrl, token)
	if err != nil {
		return nil, err
	}
	plexObj.devices = make([]*Device, 0)
	cfg = config.Get()
	go func() {
		for {
			plexObj.RefreshDevices()
			time.Sleep(time.Second * time.Duration(cfg.CastScanInterval))
		}
	}()
	return plexObj, nil
}

func Get() *Plex {
	return plexObj
}
