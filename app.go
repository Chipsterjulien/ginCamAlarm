package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

const (
	tmpfs      = "tmpfs"
	motionOnly = "motionOnly"
	start      = "start"
	stop       = "stop"
)

func main() {
	confPath := "/etc/gincamalarm/"
	confFilename := "gincamalarm"
	logFilename := "/var/log/gincamalarm/error.log"

	fd := initLogging(&logFilename)
	defer fd.Close()

	loadConfig(&confPath, &confFilename)

	startApp()
}

func createFile() {
	slurp, err := os.OpenFile(viper.GetString("server.startalarm"), os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log := logging.MustGetLogger("log")

		log.Criticalf("Unable to create \"%s\" file !", viper.GetString("server.startalarm"))
		os.Exit(1)
	}

	defer slurp.Close()
}

func getStateAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	var alarm, stream string

	out, err := exec.Command("/bin/sh", "-c", "pgrep ^motion$").Output()
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

func gpioStartStop(startStop string) {
	log := logging.MustGetLogger("log")

	gpio0List := viper.GetStringSlice("atstart.gpioto0")
	gpio1List := viper.GetStringSlice("atstart.gpioto1")

	cmdJob := make([]string, (len(gpio0List)+len(gpio1List))*2)

	log.Debugf("List of gpio to set to 0 at start: %v", gpio0List)
	log.Debugf("List of gpio to set to 1 at start: %v", gpio1List)

	counter := 0
	for _, gpio := range gpio0List {
		cmdJob[counter] = fmt.Sprintf("gpio mode %s out", gpio)
		counter++
	}

	for _, gpio := range gpio1List {
		cmdJob[counter] = fmt.Sprintf("gpio mode %s out", gpio)
		counter++
	}

	switch startStop {
	case start:
		for _, gpio := range gpio0List {
			cmdJob[counter] = fmt.Sprintf("gpio write %s 0", gpio)
			counter++
		}

		for _, gpio := range gpio1List {
			cmdJob[counter] = fmt.Sprintf("gpio write %s 1", gpio)
			counter++
		}
	case stop:
		for _, gpio := range gpio0List {
			cmdJob[counter] = fmt.Sprintf("gpio write %s 1", gpio)
			counter++
		}

		for _, gpio := range gpio1List {
			cmdJob[counter] = fmt.Sprintf("gpio write %s 0", gpio)
			counter++
		}
	}

	log.Debugf("list of jobs: %v", cmdJob)

	for _, job := range cmdJob {
		if _, err := exec.Command("/usr/bin/sh", "-c", job).Output(); err != nil {
			log.Warningf("Unable to exec \"%s\" !", job)
		}
	}
}

func isStarted() bool {
	log := logging.MustGetLogger("log")

	if _, err := os.Stat(viper.GetString("server.startalarm")); err != nil {
		log.Debugf("File %s not exist", viper.GetString("server.startalarm"))

		return false
	}

	return true
}

func removeFile() {
	if err := os.Remove(viper.GetString("server.startalarm")); err != nil {
		log := logging.MustGetLogger("log")

		log.Criticalf("Unable to remove \"%s\" file !", viper.GetString("server.startalarm"))
		os.Exit(1)
	}
}

func startApp() {
	log := logging.MustGetLogger("log")

	if viper.GetString("logtype") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	g := gin.Default()

	g.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))

	v1 := g.Group("api/v1")
	{
		v1.GET("/stateAlarm", getStateAlarm)
		v1.PUT("/startAlarm", startAlarm)
		v1.PUT("/stopAlarm", stopAlarm)
		v1.PUT("/startStream", startStream)
		v1.PUT("/stopStream", stopStream)
	}

	if isStarted() {
		restartAlarm()
	}

	log.Debugf("Port: %d", viper.GetInt("server.port"))
	if err := g.Run(":" + strconv.Itoa(viper.GetInt("server.port"))); err != nil {
		log.Criticalf("Unable to start serveur: %v", err)
		os.Exit(1)
	}
}
