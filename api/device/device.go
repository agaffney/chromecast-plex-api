package device

import (
	"github.com/agaffney/chromecast-plex-api/chromecast"
	//castplex "github.com/agaffney/chromecast-plex-api/chromecast/plex"
	"github.com/agaffney/chromecast-plex-api/plex"
	"github.com/gin-gonic/gin"
)

func ConfigureRouter(g *gin.Engine) {
	g.GET("/devices/", handleListDevices)
	g.POST("/devices/rescan", handleRescan)
	g.GET("/device/:id/", handleGetDevice)
	//g.POST("/device/:id/launch", handleLaunch)
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
	p := plex.Get()
	c.JSON(200, p.GetDevices())
}

func handleGetDevice(c *gin.Context) {
	id := c.Param("id")
	p := plex.Get()
	device := p.GetDevice(id)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, device)
}

/*
func handleLaunch(c *gin.Context) {
	id := c.Param("id")
	p := plex.Get()
	device := p.GetDevice(id)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	controller := castplex.NewController(device)
	if err := controller.Launch(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "launch triggered"})
}
*/
