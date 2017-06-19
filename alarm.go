package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

func restartAlarm() {
	log := logging.MustGetLogger("log")

	motion := viper.GetString("default.motionProgram")
	method := viper.GetString("server.method")
	camWidth := viper.GetInt("raspistill.camWidth")
	camHeight := viper.GetInt("raspistill.camHeight")
	timeLaps := viper.GetInt("raspistill.timeLaps")
	angle := viper.GetInt("raspistill.angle")
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
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -t 0 -rot %d -tl %d -w %d -h %d  -bm", "/media/tmpfs/picture.jpg", angle, timeLaps, camWidth, camHeight),
			motion,
			"mailmotion",
		}

	case motionOnly:
		cmdList = []string{
			motion,
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
		cmd := fmt.Sprintf("pgrep ^%s$", motion)
		if _, err := exec.Command("/bin/sh", "-c", cmd).Output(); err != nil {
			log.Warning("motion is not running")
			exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output()
			exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()

			return
		}

		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mailmotion$").Output(); err != nil {
			log.Warning("mailmotion is not running and there is something wrong with mailmotion")
			exec.Command("/bin/sh", "-c", fmt.Sprintf("killall -9 %s", motion)).Output()
			exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()

			return
		}
	case motionOnly:
		cmd := fmt.Sprintf("pgrep ^%s$", motion)
		if _, err := exec.Command("/bin/sh", "-c", cmd).Output(); err != nil {
			log.Warningf("%s is not running", motion)
			exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output()

			return
		}

		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mailmotion$").Output(); err != nil {
			exec.Command("/bin/sh", "-c", fmt.Sprintf("killall -9 %s", motion)).Output()
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

	motion := viper.GetString("default.motionProgram")
	method := viper.GetString("server.method")
	angle := viper.GetInt("raspistill.angle")
	camWidth := viper.GetInt("raspistill.camWidth")
	camHeight := viper.GetInt("raspistill.camHeight")
	timeLaps := viper.GetInt("raspistill.timeLaps")
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
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -t 0 -rot %d -tl %d -w %d -h %d  -bm", "/media/tmpfs/picture.jpg", angle, timeLaps, camWidth, camHeight),
			motion,
			"mailmotion",
		}

	case motionOnly:
		cmdList = []string{
			motion,
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
		if _, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("pgrep ^%s$", motion)).Output(); err != nil {
			log.Warningf("%s is not running", motion)
			exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output()
			exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()

			c.JSON(500, gin.H{"error": fmt.Sprintf("%s is not running", motion), "alarm": stop, "stream": stop, "location": viper.GetString("server.location")})

			return
		}

		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mailmotion$").Output(); err != nil {
			exec.Command("/bin/sh", "-c", fmt.Sprintf("killall -9 %s", motion)).Output()
			exec.Command("/bin/sh", "-c", "killall -9 raspistill").Output()
			log.Warning("mailmotion is not running and there is something wrong with mailmotion")

			c.JSON(500, gin.H{"error": "mailmotion is not running", "alarm": stop, "stream": stop, "location": viper.GetString("server.location")})

			return
		}
	case motionOnly:
		if _, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("pgrep ^%s$", motion)).Output(); err != nil {
			log.Warning("motion is not running")
			exec.Command("/bin/sh", "-c", "killall -9 mailmotion").Output()

			c.JSON(500, gin.H{"error": "motion is not running", "alarm": stop, "stream": stop, "location": viper.GetString("server.location")})

			return
		}

		if _, err := exec.Command("/bin/sh", "-c", "pgrep ^mailmotion$").Output(); err != nil {
			exec.Command("/bin/sh", "-c", fmt.Sprintf("killall -9 %s", motion)).Output()
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

func stopAlarm(c *gin.Context) {
	log := logging.MustGetLogger("log")
	motion := viper.GetString("default.motionProgram")
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

	if _, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("killall -9 %s", motion)).Output(); err != nil {
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
