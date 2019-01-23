package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/gin-gonic/gin"
	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

// Pour charger la page, voir ici: https://www.sigmdel.ca/michel/ha/rpi/streaming_en.html
// D'autres info intéressantes ici: https://serverfault.com/questions/788173/nginx-reverse-proxy-config-videomjpg-stream-to-use-a-single-connection-to-the

func startStream(c *gin.Context) {
	log := logging.MustGetLogger("log")

	method := viper.GetString("default.method")
	angle := viper.GetInt("raspistill.angle")
	timeLaps := viper.GetInt("raspistill.timeLaps")
	streamWidth := viper.GetInt("mjpgstreamer.streamWidth")
	streamHeight := viper.GetInt("mjpgstreamer.streamHeight")
	tmpfsPath := viper.GetString("default.tmpfsPath")
	pictureFilename := viper.GetString("default.pictureFilename")
	pictureTempfsCompletPath := path.Join(tmpfsPath, pictureFilename)
	cmdList := []string{}

	if method == streamerOnly {
		method = tmpfs
	}

	gpioStartStop(start)

	switch method {
	case tmpfs:
		if err := isPrepareTmpfsMethode(); err != nil {
			log.Critical(err)
			c.JSON(500, gin.H{"error": err})
			gpioStartStop(stop)

			return
		}

		cmdList = []string{
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -t 0 -rot %d -tl %d -w %d -h %d -bm", pictureTempfsCompletPath, angle, timeLaps, streamWidth, streamHeight),
		}

		login := viper.GetString("default.loginCam")
		password := viper.GetString("default.passwordCam")
		port := viper.GetInt("mjpgstreamer.port")
		activateIdentification := viper.GetBool("mjpgstreamer.activateIdentification")

		if activateIdentification && len(login) > 0 && len(password) > 0 {
			cmd := fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\" -p %d -c %s:%s", tmpfsPath, "picture.jpg", port, login, password)
			cmdList = append(cmdList, cmd)
		} else {
			cmd := fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\" -p %d", tmpfsPath, "picture.jpg", port)
			cmdList = append(cmdList, cmd)
		}

	case motionOnly:
		login := viper.GetString("default.loginCam")
		password := viper.GetString("default.passwordCam")
		port := viper.GetInt("mjpgstreamer.port")
		activateIdentification := viper.GetBool("mjpgstreamer.activateIdentification")

		if activateIdentification && len(login) > 0 && len(password) > 0 {
			cmd := fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\" -p %d -c %s:%s", tmpfsPath, "picture.jpg", port, login, password)
			cmdList = append(cmdList, cmd)
		} else {
			cmd := fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\" -p %d", tmpfsPath, "picture.jpg", port)
			cmdList = append(cmdList, cmd)
		}

	default:
		log.Criticalf("Unknown \"%s\" method in config file !", method)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Unknown \"%s\" method in config file !", method)})
		gpioStartStop(stop)

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

	location := viper.GetString("server.location")

	if err := isAllStarted(c, &cmdList); err != nil {
		killAll(&cmdList)
		gpioStartStop(stop)

		c.JSON(500, gin.H{"alarm": stop, "stream": stop, "location": location, "error": err})
	} else {
		gpioStartStop(start)

		streamIsStarted = true
		login := viper.GetString("default.loginCam")
		password := viper.GetString("default.passwordCam")
		port := viper.GetInt("mjpgstreamer.port")
		activateIdentification := viper.GetBool("mjpgstreamer.activateIdentification")
		sendIdentificationInJSON := viper.GetBool("mjpgstreamer.sendIdentificationInJSON")

		if activateIdentification && sendIdentificationInJSON && len(login) > 0 && len(password) > 0 {
			c.JSON(200, gin.H{"alarm": stop, "stream": start, "location": location, "login": login, "passwd": password, "port": port})
		} else {
			c.JSON(200, gin.H{"alarm": stop, "stream": start, "location": location, "port": port})
		}
	}
}

func stopStream(c *gin.Context) {
	method := viper.GetString("default.method")
	cmdList := []string{}

	if method == streamerOnly {
		method = tmpfs
	}

	switch method {
	case tmpfs:
		cmdList = []string{
			"/opt/vc/bin/raspistill",
			"mjpg_streamer",
		}
	case motionOnly:
		cmdList = []string{"mjpg_streamer"}
	}

	killAll(&cmdList)

	switch method {
	case tmpfs:
		os.Remove(viper.GetString("default.picturePathStore"))
		os.Remove(viper.GetString("default.tmpfsPath"))
	}

	gpioStartStop(stop)

	streamIsStarted = false
	c.JSON(200, gin.H{"alarm": stop, "stream": stop, "location": viper.GetString("server.location")})
}
