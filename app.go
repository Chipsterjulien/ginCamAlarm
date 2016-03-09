package main

import (
	"os/exec"
	"strconv"
	"time"
	"os"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

func GetStateAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	out, err := exec.Command("/bin/sh", "-c", "pgrep ^motion$").Output()
	if err != nil {
		log.Debug(fmt.Sprintf("Retour de la commande pgrep: %v", err))
		c.JSON(200, gin.H{"state": "stop", "location": viper.GetString("server.location")})
	} else {
		log.Debug(fmt.Sprintf("Retour de la commande pgrep: %s", out))
		c.JSON(200, gin.H{"state": "start", "location": viper.GetString("server.location")})
	}

}

func StartAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	if err := os.MkdirAll("/tmp/motion", 0744); err != nil {
		log.Warning(fmt.Sprintf("Unable to create /tmp/motion directories: %v", err))
		c.JSON(500, gin.H{"error": "mailmotion is not running", "state": "stop", "location": viper.GetString("server.location")})

		return
	}

	cmdList := []string{
		"/opt/vc/bin/raspistill -o /media/tmpfs/picture.jpg -t 0 -tl 500 -w 640 -h 480 -bm",
		"motion",
		"mailmotion",
	}

	for _, cmd := range cmdList {
		go func (cmd string) {
			m := exec.Command("/bin/sh", "-c", cmd)
			m.Start()
			m.Wait()
		}(cmd)
	}

	if _, err := exec.Command("/bin/sh", "-c", "pgrep ^motion$").Output(); err != nil {
		log.Warning("motion is not running")
		exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output()
		exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()

		c.JSON(500, gin.H{"error": "motion is not running", "state": "stop", "location": viper.GetString("server.location")})

		return
	}

	if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mailmotion$").Output(); err != nil {
		exec.Command("/bin/sh", "-c", "killall -9 motion").Output()
		exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()
		log.Warning("mailmotion is not running and there is something wrong with mailmotion")

		c.JSON(500, gin.H{"error": "mailmotion is not running", "state": "stop", "location": viper.GetString("server.location")})

		return
	}

	c.JSON(200, gin.H{"state": "start", "location": viper.GetString("server.location")})
}

func StartStream(c *gin.Context) {
	cmdList := []string{
		"/opt/vc/bin/raspistill -o /media/tmpfs/picture.jpg -t 0 -tl 500 -w 640 -h 480 -bm",
		"LD_LIBRARY_PATH=/usr/lib mjpg_streamer -i \"input_file.so -f /media/tmpfs -n picture.jpg\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\""
	}

	for _, cmd := range cmdList {
		go func (cmd string) {
			m := exec.Command("/bin/sh", "-c", cmd)
			m.Start()
			m.Wait()
		}(cmd)
	}
	
	if _, err := exec.Command("/bin/sh", "-c", "pgrep ^raspistill$").Output(); err != nil {
		exec.Command("/bin/sh", "-c", "killall -9 mjpg_stream").Output()
		c.JSON(500, gin.H{"error": "raspistill is not running", "stream": "stop", "location": viper.GetString("server.location")})

		return
	}

	if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mjpg_stream$").Output(); err != nil {
		exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()
		c.JSON(500, gin.H{"error": "mjpg_stream is not running", "stream": "stop", "location": viper.GetString("server.location")})

		return
	}

	c.JSON(200, gin.H{"stream": "start", "location": viper.GetString("server.location")})
}

func StopAlarm(c *gin.Context) {
	exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()

	if _, err := exec.Command("/bin/sh", "-c", "killall -9 motion").Output(); err != nil {
		c.JSON(500, gin.H{"error": "Unable to stop motion", "location": viper.GetString("server.location")})

		return
	}
	if _, err := exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output(); err != nil {
		c.JSON(500, gin.H{"error": "Unable to stop mailmotion", "location": viper.GetString("server.location")})

		return
	}

	if err := os.RemoveAll("/tmp/motion"); err != nil {
		c.JSON(500, gin.H{"error": "Unable to remove directory"})

		return
	}

	c.JSON(200, gin.H{"state": "stop", "location": viper.GetString("server.location")})
}

func StopStream(c *gin.Context) {
	if _, err := exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output(); err != nil {
		c.JSON(500, gin.H{"error": "Unable to stop raspistill", "location": viper.GetString("server.location")})

		return
	}
	if _, err := exec.Command("/bin/sh", "-c", "killall -9 mjpg_stream").Output(); err != nil {
		c.JSON(500, gin.H{"error": "Unable to stop mjpg_stream", "location": viper.GetString("server.location")})

		return
	}

	c.JSON(200, gin.H{"stream": "stop", "location": viper.GetString("server.location")})
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
		v1.PUT("/startAlarm", StartAlarm)
		v1.PUT("/stopAlarm", StopAlarm)
		v1.PUT("/startStream", StartStream)
		v1.PUT("/stopStream", StopStream)
	}

	log.Debug(fmt.Sprintf("Port: %d", viper.GetInt("server.port")))
	if err := g.Run(":" + strconv.Itoa(viper.GetInt("server.port"))); err != nil {
		log.Critical(fmt.Sprintf("Unable to start serveur: %v", err))
		os.Exit(1)
	}
}

func main() {
	confPath := "/etc/gincamalarm/"
	confFilename := "gincamalarm"
	logFilename := "/var/log/gincamalarm/error.log"

	// Si je veux savoir si un processus est déjà lancé, il faut que j'utilise pgrep pour ensuite kill la bonne appli avec son numéro

	// confPath := "cfg"
	// confFilename := "gincamalarm"
	// logFilename := "error.log"

	fd := initLogging(&logFilename)
	defer fd.Close()

	loadConfig(&confPath, &confFilename)

	startApp()
}