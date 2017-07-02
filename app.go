package main

import (
	"os"
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
