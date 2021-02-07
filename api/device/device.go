package device

import (
	"github.com/agaffney/chromecast-plex-api/chromecast"
	"github.com/agaffney/chromecast-plex-api/chromecast/plex"
	"github.com/gin-gonic/gin"
)

func ConfigureRouter(g *gin.Engine) {
	g.GET("/devices/", handleListDevices)
	g.POST("/devices/rescan", handleRescan)
	g.GET("/device/:uuid/", handleGetDevice)
	g.POST("/device/:uuid/launch", handleLaunch)
	g.POST("/device/:uuid/update", handleUpdate)
	g.POST("/device/:uuid/reset", handleReset)
	configureRouterPlayback(g)
	configureRouterVolume(g)
}

func handleRescan(c *gin.Context) {
	go func() {
		chromecast.Scan()
	}()
	c.JSON(200, gin.H{"message": "rescan triggered"})
}

func handleListDevices(c *gin.Context) {
	c.JSON(200, chromecast.GetDevices())
}

func handleGetDevice(c *gin.Context) {
	uuid := c.Param("uuid")
	device := chromecast.GetDevice(uuid)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, device)
}

func handleLaunch(c *gin.Context) {
	uuid := c.Param("uuid")
	device := chromecast.GetDevice(uuid)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	controller := plex.NewController(device)
	if err := controller.Launch(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "launch triggered"})
}

func handleUpdate(c *gin.Context) {
	uuid := c.Param("uuid")
	device := chromecast.GetDevice(uuid)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if err := device.Update(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "update triggered"})
}

func handleReset(c *gin.Context) {
	uuid := c.Param("uuid")
	device := chromecast.GetDevice(uuid)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if err := device.Reset(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "reset triggered"})
}
