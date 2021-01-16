package api

import (
	"fmt"
	"github.com/agaffney/chromecast-plex-api/config"
	"github.com/gin-gonic/gin"
)

func Start() {
	config := config.Get()
	if config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	router.Run(fmt.Sprintf("%s:%d", config.Address, config.Port))
}
