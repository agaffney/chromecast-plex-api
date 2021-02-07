package device

import (
	"github.com/agaffney/chromecast-plex-api/chromecast"
	"github.com/gin-gonic/gin"
)

func configureRouterVolume(g *gin.Engine) {
	// Volume
	g.POST("/device/:uuid/volume/mute", handleVolumeMute)
	g.POST("/device/:uuid/volume/unmute", handleVolumeUnmute)
	g.POST("/device/:uuid/volume/up", handleVolumeUp)
	g.POST("/device/:uuid/volume/down", handleVolumeDown)
	g.POST("/device/:uuid/volume/setLevel", handleVolumeSetLevel)
}

func handleVolumeMute(c *gin.Context) {
	uuid := c.Param("uuid")
	device := chromecast.GetDevice(uuid)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if err := device.SetVolumeMuted(true); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "volume muted"})
}

func handleVolumeUnmute(c *gin.Context) {
	uuid := c.Param("uuid")
	device := chromecast.GetDevice(uuid)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if err := device.SetVolumeMuted(false); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "volume unmuted"})
}

func handleVolumeUp(c *gin.Context) {
	uuid := c.Param("uuid")
	device := chromecast.GetDevice(uuid)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	volume := device.GetVolume()
	if err := device.SetVolumeLevel(volume.Level + 0.1); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "volume raised"})
}

func handleVolumeDown(c *gin.Context) {
	uuid := c.Param("uuid")
	device := chromecast.GetDevice(uuid)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	volume := device.GetVolume()
	if err := device.SetVolumeLevel(volume.Level - 0.1); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "volume lowered"})
}

type volumeSetLevelRequest struct {
	Level float32 `json:"level"`
}

func handleVolumeSetLevel(c *gin.Context) {
	uuid := c.Param("uuid")
	device := chromecast.GetDevice(uuid)
	if device == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	var request volumeSetLevelRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := device.SetVolumeLevel(request.Level); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "volume level set"})
}
