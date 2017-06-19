package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

func startStream(c *gin.Context) {
	log := logging.MustGetLogger("log")

	method := viper.GetString("server.method")
	angle := viper.GetInt("raspistill.angle")
	camWidth := viper.GetInt("raspistill.camWidth")
	camHeight := viper.GetInt("raspistill.camHeight")
	timeLaps := viper.GetInt("raspistill.timeLaps")
	streamWidth := viper.GetInt("mjpgstreamer.streamWidth")
	streamHeight := viper.GetInt("mjpgstreamer.streamHeight")
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
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -t 0 -rot %d -tl %d -w %d -h %d -bm", filename, angle, timeLaps, camWidth, camHeight),
			// raspistill -o /media/tmpfs/picture.jpg -t 0 -tl 750 -w 640 -h 480 -bm
			fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\"", tmpfsPath, "picture.jpg"),
			//           LD_LIBRARY_PATH=/usr/lib mjpg_streamer -i "input_file.so -f /media/tmpfs -n picture.jpg -r 320x240" -o "output_http.so -w /usr/share/mjpg-streamer/www/"
		}
	case motionOnly:
		cmdList = []string{
			fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -i \"input_uvc.so -y -r %dx%d -f 1\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\"", streamWidth, streamHeight),
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
