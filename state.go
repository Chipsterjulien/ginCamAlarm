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

func getRealExec(cmd *string) {
	*cmd = strings.Replace(*cmd, "LD_LIBRARY_PATH=/usr/lib ", "", -1)
	*cmd = path.Base(strings.Split(*cmd, " ")[0])
}

func isInList(progList *[]string, prog *string) bool {
	for _, p := range *progList {
		if p == *prog {
			return true
		}
	}

	return false
}

func getStateAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	alarm := stop
	stream := stop
	progList := []string{}
	programList := viper.GetStringSlice("default.motionProgram")
	method := viper.GetString("default.method")

	if isStarted() {
		for _, prog := range programList {
			getRealExec(&prog)
			if !isInList(&progList, &prog) {
				progList = append(progList, prog)
			}
		}

		switch method {
		case tmpfs:
			anotherList := []string{"raspistill", "mailmotion"}
			progList = append(progList, anotherList...)
		case motionOnly:
			progList = append(progList, "mailmotion")
		}

		for _, prog := range progList {
			// log.Debugf("Valeur de prog dans getState: %s", prog)
			cmdStr := fmt.Sprintf("pgrep ^%s$", prog)
			out, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
			if err != nil {
				// log.Debugf("Dans la fonction state, je passe dans le err avec la cmd: %s", cmdStr)
				log.Debugf("Retour de la commande pgrep: %s", err)
			} else {
				// log.Debugf("Dans la fonction state, je ne passe pas dans le err avec la cmd: %s", cmdStr)
				log.Debugf("Retour de la commande pgrep: %s", out)
				alarm = start
				break
			}
		}
	} else {
		out, err := exec.Command("/bin/sh", "-c", "pgrep ^mjpg_streamer$").Output()
		if err != nil {
			log.Debugf("Retour de la commande pgrep: %s", err)
		} else {
			log.Debugf("Retour de la commande pgrep: %s", out)
			stream = start
		}
	}

	c.JSON(200, gin.H{"alarm": alarm, "stream": stream, "location": viper.GetString("server.location")})
}
