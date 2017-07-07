package main

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

func getStateAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	alarm := stop
	stream := stop

	realExec := strings.Split(path.Base(viper.GetString("default.motionProgram")), " ")[0]
	cmdStr := fmt.Sprintf("pgrep ^%s$", realExec)

	out, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		log.Debugf("Retour de la commande pgrep: %s", err)
	} else {
		log.Debugf("Retour de la commande pgrep: %s", out)
		alarm = start
	}

	out, err = exec.Command("/bin/sh", "-c", "pgrep ^mjpg_streamer$").Output()
	if err != nil {
		log.Debugf("Retour de la commande pgrep: %s", err)
	} else {
		log.Debugf("Retour de la commande pgrep: %s", out)
		if alarm != start {
			stream = start
		}
	}

	c.JSON(200, gin.H{"alarm": alarm, "stream": stream, "location": viper.GetString("server.location")})
}
