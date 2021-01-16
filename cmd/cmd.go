package cmd

import (
	"github.com/agaffney/chromecast-plex-api/api"
	"github.com/agaffney/chromecast-plex-api/chromecast"
	"github.com/agaffney/chromecast-plex-api/config"
	"log"
)

func Start() {
	_, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}
	chromecast.StartScanning()
	api.Start()
}
