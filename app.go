package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
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
		log.Debugf("File %s not exist", viper.GetString("server.sartalarm"))

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

func restartAlarm() {
	log := logging.MustGetLogger("log")

	method := viper.GetString("server.method")
	angle := viper.GetInt("server.angle")
	cmdList := []string{}

	switch method {
	case tmpfs:
		if _, err := os.Stat("/media/tmpfs/picture.jpg"); err == nil {
			if er := os.Remove("/media/tmpfs/picture.jpg"); er != nil {

				return
			}
		}

		if err := os.MkdirAll("/tmp/motion/send", 0744); err != nil {
			log.Warningf("Unable to create /tmp/motion directories: %v", err)

			return
		}

		cmdList = []string{
			fmt.Sprintf("/opt/vc/bin/raspistill -o /media/tmpfs/picture.jpg -t 0 -rot %d -tl 459 -w 640 -h 480 -bm", angle),
			"motion",
			"mailmotion",
		}

	case motionOnly:
		cmdList = []string{
			"motion",
			"mailmotion",
		}
	default:
		log.Criticalf("Unknown \"%s\" method in config file !", method)
		os.Exit(1)
	}

	for _, cmd := range cmdList {
		go func(cmd string) {
			log.Infof("Launch command: \"%s\"", cmd)
			m := exec.Command("/bin/sh", "-c", cmd)
			m.Start()
			m.Wait()
		}(cmd)
	}

	switch method {
	case tmpfs:
		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^motion$").Output(); err != nil {
			log.Warning("motion is not running")
			exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output()
			exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()

			return
		}

		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mailmotion$").Output(); err != nil {
			exec.Command("/bin/sh", "-c", "killall -9 motion").Output()
			exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()
			log.Warning("mailmotion is not running and there is something wrong with mailmotion")

			return
		}
	case motionOnly:
		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^motion$").Output(); err != nil {
			log.Warning("motion is not running")
			exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output()

			return
		}

		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mailmotion$").Output(); err != nil {
			exec.Command("/bin/sh", "-c", "killall -9 motion").Output()
			log.Warning("mailmotion is not running and there is something wrong with mailmotion")

			return
		}
	default:
		log.Criticalf("Unknown \"%s\" method in config file !", method)

		os.Exit(1)
	}

	gpioStartStop(start)
}

func startAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")

	method := viper.GetString("server.method")
	angle := viper.GetInt("server.angle")
	cmdList := []string{}

	switch method {
	case tmpfs:
		if _, err := os.Stat("/media/tmpfs/picture.jpg"); err == nil {
			if er := os.Remove("/media/tmpfs/picture.jpg"); er != nil {
				c.JSON(500, gin.H{"error": "Unable to start stream since unable to remove picture.jpg in tmpfs"})

				return
			}
		}

		if err := os.MkdirAll("/tmp/motion/send", 0744); err != nil {
			log.Warningf("Unable to create /tmp/motion directories: %v", err)
			c.JSON(500, gin.H{"error": "mailmotion is not running", "alarm": stop, "location": viper.GetString("server.location")})

			return
		}

		cmdList = []string{
			fmt.Sprintf("/opt/vc/bin/raspistill -o /media/tmpfs/picture.jpg -t 0 -rot %d -tl 459 -w 640 -h 480 -bm", angle),
			"motion",
			"mailmotion",
		}

	case motionOnly:
		cmdList = []string{
			"motion",
			"mailmotion",
		}
	default:
		log.Criticalf("Unknown \"%s\" method in config file !", method)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Unknown \"%s\" method in config file !", method)})
		os.Exit(1)
	}

	for _, cmd := range cmdList {
		go func(cmd string) {
			log.Infof("Launch command: \"%s\"", cmd)
			m := exec.Command("/bin/sh", "-c", cmd)
			m.Start()
			m.Wait()
		}(cmd)
	}

	switch method {
	case tmpfs:
		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^motion$").Output(); err != nil {
			log.Warning("motion is not running")
			exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output()
			exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()

			c.JSON(500, gin.H{"error": "motion is not running", "alarm": stop, "stream": stop, "location": viper.GetString("server.location")})

			return
		}

		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mailmotion$").Output(); err != nil {
			exec.Command("/bin/sh", "-c", "killall -9 motion").Output()
			exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()
			log.Warning("mailmotion is not running and there is something wrong with mailmotion")

			c.JSON(500, gin.H{"error": "mailmotion is not running", "alarm": stop, "stream": stop, "location": viper.GetString("server.location")})

			return
		}
	case motionOnly:
		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^motion$").Output(); err != nil {
			log.Warning("motion is not running")
			exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output()

			c.JSON(500, gin.H{"error": "motion is not running", "alarm": stop, "stream": stop, "location": viper.GetString("server.location")})

			return
		}

		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mailmotion$").Output(); err != nil {
			exec.Command("/bin/sh", "-c", "killall -9 motion").Output()
			log.Warning("mailmotion is not running and there is something wrong with mailmotion")

			c.JSON(500, gin.H{"error": "mailmotion is not running", "alarm": stop, "stream": stop, "location": viper.GetString("server.location")})

			return
		}
	default:
		log.Criticalf("Unknown \"%s\" method in config file !", method)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Unknown \"%s\" method in config file !", method)})
		os.Exit(1)
	}

	gpioStartStop(start)

	createFile()

	c.JSON(200, gin.H{"alarm": start, "stream": stop, "location": viper.GetString("server.location")})
}

func startStream(c *gin.Context) {
	log := logging.MustGetLogger("log")

	method := viper.GetString("server.method")
	angle := viper.GetInt("server.angle")
	cmdList := []string{}

	switch method {
	case tmpfs:
		tmpfsPath := viper.GetString("server.tmpfsPath")
		filename := path.Join(tmpfsPath, "picture.jpg")

		if _, err := os.Stat(filename); err == nil {
			if er := os.Remove(filename); er != nil {
				c.JSON(500, gin.H{"error": fmt.Sprintf("Unable to start stream since unable to remove picture.jpg in tmpfs: %s", er)})

				return
			}
		}

		cmdList = []string{
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -rot %d -t 0 -q 8 -tl 1000 -w 320 -h 240 -bm", filename, angle),
			fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -i \"input_file.so -f %s -n picture.jpg\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\"", tmpfsPath),
		}
	case motionOnly:
		cmdList = []string{
			"LD_LIBRARY_PATH=/usr/lib mjpg_streamer -i \"input_uvc.so -y -r 320x240 -f 1\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\"",
		}
	default:
		log.Criticalf("Unknown \"%s\" method in config file !", method)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Unknown \"%s\" method in config file !", method)})
		os.Exit(1)
	}

	for _, cmd := range cmdList {
		go func(cmd string) {
			log.Infof("Launch command: \"%s\"", cmd)
			m := exec.Command("/bin/sh", "-c", cmd)
			m.Start()
			m.Wait()
		}(cmd)
	}

	switch method {
	case tmpfs:
		if out, err := exec.Command("/bin/sh", "-c", "pgrep ^raspistill$").Output(); err != nil {
			log.Debugf("Sortie: %s", string(out))

			exec.Command("/bin/sh", "-c", "killall -9 mjpg_streamer").Output()
			c.JSON(500, gin.H{"error": "raspistill is not running", "alarm": stop, "stream": stop, "location": viper.GetString("server.location")})

			return
		}

		if out, err := exec.Command("/bin/sh", "-c", "pgrep ^mjpg_streamer$").Output(); err != nil {
			log.Debugf("Sortie: %s", string(out))
			log.Debug("mjpg_streamer a merdé")

			exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()
			c.JSON(500, gin.H{"error": "mjpg_stream is not running", "alarm": stop, "stream": stop, "location": viper.GetString("server.location")})

			return
		}
	case motionOnly:
		if out, err := exec.Command("/bin/sh", "-c", "pgrep ^mjpg_streamer$").Output(); err != nil {
			log.Debugf("Sortie: %s", string(out))
			log.Debug("mjpg_streamer a merdé")

			c.JSON(500, gin.H{"error": "mjpg_stream is not running", "alarm": stop, "stream": stop, "location": viper.GetString("server.location")})

			return
		}
	}

	gpioStartStop(start)

	c.JSON(200, gin.H{"alarm": stop, "stream": start, "location": viper.GetString("server.location")})
}

func stopAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")
	method := viper.GetString("server.method")

	switch method {
	case tmpfs:
		exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()
	case motionOnly:
	default:
		log.Criticalf("Unknown \"%s\" method in config file !", method)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Unknown \"%s\" method in config file !", method)})
		os.Exit(1)
	}

	if _, err := exec.Command("/bin/sh", "-c", "killall -9 motion").Output(); err != nil {
		c.JSON(500, gin.H{"error": "Unable to stop motion", "location": viper.GetString("server.location")})

		return
	}
	if _, err := exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output(); err != nil {
		c.JSON(500, gin.H{"error": "Unable to stop mailmotion", "location": viper.GetString("server.location")})

		return
	}

	switch method {
	case tmpfs:
		if err := os.RemoveAll("/tmp/motion"); err != nil {
			c.JSON(500, gin.H{"error": "Unable to remove directory"})

			return
		}

		if err := os.Remove("/media/tmpfs/picture.jpg"); err != nil {
			c.JSON(500, gin.H{"error": "Unable to remove picture.jpg in tmpfs"})

			return
		}
	case motionOnly:
	}

	gpioStartStop(stop)

	removeFile()

	c.JSON(200, gin.H{"alarm": stop, "stream": stop, "location": viper.GetString("server.location")})
}

func stopStream(c *gin.Context) {
	// log := logging.MustGetLogger("log")
	method := viper.GetString("server.method")

	switch method {
	case tmpfs:
		if _, err := exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output(); err != nil {
			c.JSON(500, gin.H{"error": "Unable to stop raspistill", "location": viper.GetString("server.location")})

			return
		}
	case motionOnly:
	default:
	}

	if _, err := exec.Command("/bin/sh", "-c", "killall -9 mjpg_streamer").Output(); err != nil {
		c.JSON(500, gin.H{"error": "Unable to stop mjpg_stream", "location": viper.GetString("server.location")})

		return
	}

	switch method {
	case tmpfs:
		if err := os.Remove("/media/tmpfs/picture.jpg"); err != nil {
			c.JSON(500, gin.H{"error": "Unable to remove picture.jpg in tmpfs"})

			return
		}
	case motionOnly:
	}

	gpioStartStop(stop)

	c.JSON(200, gin.H{"alarm": stop, "stream": stop, "location": viper.GetString("server.location")})
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
