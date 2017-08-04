package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func getStateAlarm(c *gin.Context) {
	// log := logging.MustGetLogger("log")

	alarm := stop
	stream := stop
	// progList := []string{}
	// programList := viper.GetStringSlice("default.motionProgram")
	// method := viper.GetString("default.method")
	//
	// if isStarted() {
	// 	for _, prog := range programList {
	// 		getRealExec(&prog)
	// 		if !isInList(&progList, &prog) {
	// 			progList = append(progList, prog)
	// 		}
	// 	}
	//
	// 	switch method {
	// 	case tmpfs:
	// 		anotherList := []string{"raspistill", "mailmotion"}
	// 		progList = append(progList, anotherList...)
	// 	case motionOnly:
	// 		progList = append(progList, "mailmotion")
	// 	}
	//
	// 	for _, prog := range progList {
	// 		// log.Debugf("Valeur de prog dans getState: %s", prog)
	// 		cmdStr := fmt.Sprintf("pgrep ^%s$", prog)
	// 		out, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	// 		if err != nil {
	// 			// log.Debugf("Dans la fonction state, je passe dans le err avec la cmd: %s", cmdStr)
	// 			log.Debugf("Retour de la commande pgrep: %s", err)
	// 		} else {
	// 			// log.Debugf("Dans la fonction state, je ne passe pas dans le err avec la cmd: %s", cmdStr)
	// 			log.Debugf("Retour de la commande pgrep: %s", out)
	// 			alarm = start
	// 			break
	// 		}
	// 	}
	// } else {
	// 	out, err := exec.Command("/bin/sh", "-c", "pgrep ^mjpg_streamer$").Output()
	// 	if err != nil {
	// 		log.Debugf("Retour de la commande pgrep: %s", err)
	// 	} else {
	// 		log.Debugf("Retour de la commande pgrep: %s", out)
	// 		stream = start
	// 	}
	// }

	if alarmIsStarted {
		alarm = start
	}
	if streamIsStarted {
		stream = start
	}

	c.JSON(200, gin.H{"alarm": alarm, "stream": stream, "location": viper.GetString("server.location")})
}
