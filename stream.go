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

	switch method {
	case tmpfs:
		if err := isPrepareTmpfsMethode(); err != nil {
			log.Critical(err)
			c.JSON(500, gin.H{"error": err})

			return
		}

		cmdList = []string{
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -t 0 -rot %d -tl %d -w %d -h %d -bm", pictureTempfsCompletPath, angle, timeLaps, streamWidth, streamHeight),
			fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\"", tmpfsPath, pictureFilename),
		}

	case motionOnly:
		cmdList = []string{
			fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_uvc.so -y -r %dx%d -f 1\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\"", streamWidth, streamHeight),
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

	if err := isAllStarted(c, &cmdList); err != nil {
		killAll(&cmdList)

		c.JSON(500, gin.H{"alarm": stop, "stream": stop, "location": viper.GetString("server.location")})
	} else {
		gpioStartStop(start)

		c.JSON(200, gin.H{"alarm": stop, "stream": start, "location": viper.GetString("server.location")})
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

	c.JSON(200, gin.H{"alarm": stop, "stream": stop, "location": viper.GetString("server.location")})
}
