package main

import (
	"fmt"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

func getStateAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	var alarm, stream string
	motion := viper.GetString("default.motionProgram")

	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("pgrep ^%s$", motion)).Output()
	if err != nil {
		log.Debugf("Retour de la commande pgrep: %v", err)
		alarm = stop
	} else {
		log.Debugf("Retour de la commande pgrep: %s", out)
		alarm = start
	}

	out, err = exec.Command("/bin/sh", "-c", "pgrep ^mjpg_streamer$").Output()
	if err != nil {
		log.Debugf("Retour de la commande pgrep: %v", err)
		stream = stop
	} else {
		log.Debugf("Retour de la commande pgrep: %s", out)
		stream = start
	}

	c.JSON(200, gin.H{"alarm": alarm, "stream": stream, "location": viper.GetString("server.location")})
}
