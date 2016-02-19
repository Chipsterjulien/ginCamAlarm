package main

import (
	"os/exec"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

func GetStateAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	out, err := exec.Command("/bin/sh", "-c", "pgrep ^motion$").Output()
	if err != nil {
		log.Debug("Retour de la commande pgrep: %v", err)
		// renvoyer l'état
		c.JSON(200, gin.H{"state": "stop", "location": viper.GetString("server.location")})
	} else {
		log.Debug("Retour de la commande pgrep: %s", out)
		// renvoyer l'état
		c.JSON(200, gin.H{"state": "start", "location": viper.GetString("server.location")})
	}

}

func TakePicture(c *gin.Context) {
}

func StartAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	m := exec.Command("/bin/sh", "-c", "motion&")
	m.Start()
	m.Wait()
	m = exec.Command("/bin/sh", "-c", "mailmotion&")
	m.Start()
	m.Wait()

	if _, err := exec.Command("/bin/sh", "-c", "pgrep ^motion$").Output(); err != nil {
		log.Warning("motion is not running")

		c.JSON(500, gin.H{"error": "motion is not running", "state": "stop", "location": viper.GetString("server.location")})

		return
	}

	if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mailmotion$").Output(); err != nil {
		exec.Command("/bin/sh", "-c", "killall -9 motion").Output()
		log.Warning("mailmotion is not running and there is something wrong with mailmotion")

		c.JSON(500, gin.H{"error": "mailmotion is not running", "state": "stop", "location": viper.GetString("server.location")})

		return
	}

	c.JSON(200, gin.H{"state": "start", "location": viper.GetString("server.location")})
}

func StopAlarm(c *gin.Context) {
	// log := logging.MustGetLogger("log")

	if _, err := exec.Command("/bin/sh", "-c", "killall -9 motion").Output(); err != nil {
		c.JSON(500, gin.H{"error": "Unable to stop motion", "location": viper.GetString("server.location")})

		return
	}
	if _, err := exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output(); err != nil {
		c.JSON(500, gin.H{"error": "Unable to stop mailmotion", "location": viper.GetString("server.location")})

		return
	}

	c.JSON(200, gin.H{"state": "stop", "location": viper.GetString("server.location")})
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
		v1.GET("/stateAlarm", GetStateAlarm)
		v1.GET("/takePicture", TakePicture)
		v1.PUT("/startAlarm", StartAlarm)
		v1.PUT("/stopAlarm", StopAlarm)
	}

	log.Debug("Port: %d", viper.GetInt("server.port"))
	g.Run(":" + strconv.Itoa(viper.GetInt("server.port")))
}

func main() {
	confPath := "/etc/gincamalarm/"
	confFilename := "gincamalarm"
	logFilename := "/var/log/gincamalarm/error.log"

	// confPath := "cfg"
	// confFilename := "gincamalarm"
	// logFilename := "error.log"

	fd := initLogging(&logFilename)
	defer fd.Close()

	loadConfig(&confPath, &confFilename)

	startApp()
}
