package api

import (
	"github.com/agaffney/chromecast-plex-api/chromecast"
	"github.com/gin-gonic/gin"
)

func handleRescan(c *gin.Context) {
	chromecast.Scan()
	c.JSON(200, gin.H{"message": "rescan triggered"})
}

func handleListDevices(c *gin.Context) {
	c.JSON(200, chromecast.GetDevices())
}

func handleGetDevice(c *gin.Context) {
	uuid := c.Param("uuid")
	devices := chromecast.GetDevices()
	for _, device := range devices {
		if device.UUID == uuid {
			c.JSON(200, device)
			return
		}
	}
	c.JSON(404, gin.H{"error": "not found"})
}
