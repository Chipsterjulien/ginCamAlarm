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

func restartCameraTime() {
	if viper.GetBool("default.restartCamTime") {
		for {
			if alarmIsStarted {
				restartCamTime := viper.GetInt("default.restartCamTime")
				time.Sleep(time.Minute * time.Duration(restartCamTime))

				stopAlarmWithoutGinContext()
				startAlarmWithoutGinContext()
			} else {
				time.Sleep(time.Second * 10)
			}
		}
	}
}

func startApp() {
	log := logging.MustGetLogger("log")

	if viper.GetString("logtype") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	g := gin.Default()

	g.Use(cors.Middleware(
		cors.Config{
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
		v1.GET("/startAlarm", startAlarm)
		v1.GET("/stopAlarm", stopAlarm)
		v1.GET("/startStream", startStream)
		v1.GET("/stopStream", stopStream)
	}

	alwaysStartAfterAReboot := viper.GetBool("default.alwaysStartAfterAReboot")

	if alwaysStartAfterAReboot || isStarted() {
		restartAlarm()
		alarmIsStarted = true
	}

	log.Debugf("Port: %d", viper.GetInt("server.port"))
	if err := g.Run(":" + strconv.Itoa(viper.GetInt("server.port"))); err != nil {
		log.Criticalf("Unable to start serveur: %v", err)
		os.Exit(1)
	}
}

func main() {
	confPath := "/etc/gincamalarm/"
	confFilename := "gincamalarm"
	logFilename := "/var/log/gincamalarm/error.log"

	// confPath := "cfg/"
	// confFilename := "gincamalarm_sample"
	// logFilename := "errors.log"

	fd := initLogging(&logFilename)
	defer fd.Close()

	loadConfig(&confPath, &confFilename)

	go restartCameraTime()
	startApp()
}
