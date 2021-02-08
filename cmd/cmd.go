package cmd

import (
	"github.com/agaffney/chromecast-plex-api/api"
	"github.com/agaffney/chromecast-plex-api/chromecast"
	"github.com/agaffney/chromecast-plex-api/config"
	"github.com/agaffney/chromecast-plex-api/plex"
	"log"
)

func Start() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}
	_, err = plex.New(cfg.PlexUrl, cfg.PlexToken)
	if err != nil {
		log.Fatalf("failed to connect to Plex: %s", err)
	}
	chromecast.StartScanning()
	api.Start()
}
