package device

import (
	"github.com/agaffney/chromecast-plex-api/chromecast"
	"github.com/agaffney/chromecast-plex-api/chromecast/plex"
	"github.com/gin-gonic/gin"
)

func configureRouterPlayback(g *gin.Engine) {
	g.POST("/device/:uuid/playback/next", handlePlaybackNext)
}

func handlePlaybackNext(c *gin.Context) {
	uuid := c.Param("uuid")
	device := chromecast.GetDevice(uuid)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	controller := plex.NewController(device)
	if err := controller.Next(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "next triggered"})
}
