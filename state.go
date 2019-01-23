package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func getStateAlarm(c *gin.Context) {
	alarm := stop
	stream := stop

	if alarmIsStarted {
		alarm = start
	}
	if streamIsStarted {
		stream = start
	}

	c.JSON(200, gin.H{"alarm": alarm, "stream": stream, "location": viper.GetString("server.location")})
}
