package main

import (
	"strings"
	"strconv"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

func isAuthorized(clientIP string) bool {
	log := logging.MustGetLogger("log")

	if viper.GetString("logtype") == "debug" {
		for _, ip := range(viper.GetStringSlice("authorized_ip.ip")) {
			log.Debug("IP autorisée: %s", ip)
		}
	}
	for _, ip := range(viper.GetStringSlice("authorized_ip.ip")) {
		if clientIP == ip {
			log.Debug("isAuthorized retourne: true")
			return true
		}
	}

	log.Debug("isAuthorized retourne: false")
	return false
}

func GetStateAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	log.Debug("IP du client: %s", c.ClientIP())
	clientIP := strings.Split(c.ClientIP(), ":")[0]
	log.Debug("IP du client découpée: %s", clientIP)

	if isAuthorized(clientIP) {
		// chercher l'état de la caméra
		out, err := exec.Command("/bin/sh", "-c", "pgrep ^motion$").Output()
		if err != nil {
			log.Debug("Retour de la commande pgrep: %v", err)
			// renvoyer l'état
			c.JSON(200, gin.H{"state": "stop"})
		} else {
			log.Debug("Retour de la commande pgrep: %s", out)
			// renvoyer l'état
			c.JSON(200, gin.H{"state": "start"})
		}
	} else {
		// renvoyer une erreur json
		c.JSON(401, gin.H{"error": "Unauthorized access"})
	}
}

func StartAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	log.Debug("IP du client: %s", c.ClientIP())
	clientIP := strings.Split(c.ClientIP(), ":")[0]
	log.Debug("IP du client découpée: %s", clientIP)

	if !isAuthorized(clientIP) {
		// renvoyer une erreur json
		c.JSON(401, gin.H{"error": "Unauthorized access"})

		return
	}

	exec.Command("/bin/sh", "-c", "motion &").Output()
	exec.Command("/bin/sh", "-c", "mailmotiond &").Output()

	if _, err := exec.Command("/bin/sh", "-c", "pgrep ^motion$").Output(); err != nil {
		c.JSON(500, gin.H{"error": "motion is not running"})

		return
	}
	
	if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mailmotiond$").Output(); err != nil {
		exec.Command("/bin/sh", "-c", "killall -9 motion").Output()
		c.JSON(500, gin.H{"error": "mailmotiond is not running"})

		return
	}

	c.JSON(200, gin.H{"state": "start"})
}

func StopAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	log.Debug("IP du client: %s", c.ClientIP())
	clientIP := strings.Split(c.ClientIP(), ":")[0]
	log.Debug("IP du client découpée: %s", clientIP)

	if !isAuthorized(clientIP) {
		// renvoyer une erreur json
		c.JSON(401, gin.H{"error": "Unauthorized access"})

		return
	}

	if _, err := exec.Command("/bin/sh", "-c", "killall -9 motion").Output(); err != nil {
		c.JSON(500, gin.H{"error": "Unable to stop motion"})

		return
	}
	if _, err := exec.Command("/bin/sh", "-c", "killall -9 mailmotiond").Output(); err != nil {
		c.JSON(500, gin.H{"error": "Unable to stop mailmotiond"})

		return
	}

	c.JSON(200, gin.H{"state": "stop"})
}

func startApp() {
	log := logging.MustGetLogger("log")

	if viper.GetString("logtype") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	g := gin.Default()
	v1 := g.Group("api/v1")
	{
		v1.GET("/stateAlarm", GetStateAlarm)
		v1.PUT("/startAlarm", StartAlarm)
		v1.PUT("/stopAlarm", StopAlarm)
	}

	log.Debug("Port: %d", viper.GetInt("server.port"))
	g.Run(":" + strconv.Itoa(viper.GetInt("server.port")))
}

func main() {
	confPath := "/etc/ginCamAlarm/"
	confFilename := "ginCamAlarm"
	logFilename := "/var/log/ginCamAlarm/error.log"

	// confPath := "cfg"
	// confFilename := "ginCamAlarm"
	// logFilename := "error.log"

	fd := initLogging(&logFilename)
	defer fd.Close()

	loadConfig(&confPath, &confFilename)

	startApp()
}