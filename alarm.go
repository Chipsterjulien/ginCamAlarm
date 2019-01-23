package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

func isAllStarted(c *gin.Context, cmdList *[]string) error {
	log := logging.MustGetLogger("log")

	for _, cmd := range *cmdList {
		getRealExec(&cmd)

		cmdStr := fmt.Sprintf("pgrep ^%s$", cmd)

		log.Debugf("Cmd to exec: %s", cmdStr)

		if _, err := exec.Command("/bin/sh", "-c", cmdStr).Output(); err != nil {
			errorStr := fmt.Sprintf("\"%s\" is not running !", cmd)
			log.Warning(errorStr)

			return errors.New(errorStr)
		}
	}

	return nil
}

func isAllStartedWithoutGinContext(cmdList *[]string) error {
	log := logging.MustGetLogger("log")

	for _, cmd := range *cmdList {
		getRealExec(&cmd)
		cmdStr := fmt.Sprintf("pgrep ^%s$", cmd)

		log.Debugf("Cmd to exec: %s", cmdStr)

		if _, err := exec.Command("/bin/sh", "-c", cmdStr).Output(); err != nil {
			errorStr := fmt.Sprintf("\"%s\" is not running !", cmd)

			log.Warning(errorStr)

			return errors.New(errorStr)
		}
	}

	return nil
}

func isPrepareTmpfsMethode() error {
	tmpfs := viper.GetString("default.tmpfsPath")
	pictureTempfsCompletPath := path.Join(tmpfs, viper.GetString("default.pictureFilename"))
	pictureStoreCompletPath := path.Join(viper.GetString("default.picturePathStore"), viper.GetString("default.dirSendName"))

	// check if picture file exist
	if _, err := os.Stat(pictureTempfsCompletPath); err == nil {
		if er := os.Remove(pictureTempfsCompletPath); er != nil {
			errorStr := fmt.Sprintf("Unable to remove \"%s\" in tmpfs method: %s", pictureTempfsCompletPath, err)

			return errors.New(errorStr)
		}
	} else {
		// try to create a file to test write right
		randFilename := randStringBytesMaskImprSrc(10)
		randFilenameFullPath := path.Join(tmpfs, randFilename)

		fn, err := os.Create(randFilenameFullPath)
		if err != nil {
			errorStr := fmt.Sprintf("Unable to create a random file in \"%s\": %s", randFilenameFullPath, err)

			return errors.New(errorStr)
		}

		fn.Close()

		if err = os.Remove(randFilenameFullPath); err != nil {
			errorStr := fmt.Sprintf("Unable to remove \"%s\": %s", randFilenameFullPath, err)

			return errors.New(errorStr)
		}

		if err := os.MkdirAll(pictureStoreCompletPath, 0744); err != nil {
			errorStr := fmt.Sprintf("Unable to create \"%s\" directories: %s", pictureStoreCompletPath, err)

			return errors.New(errorStr)
		}
	}

	return nil
}

func restartAlarm() {
	log := logging.MustGetLogger("log")

	method := viper.GetString("default.method")
	motion := viper.GetStringSlice("default.motionProgram")
	angle := viper.GetInt("raspistill.angle")
	camWidth := viper.GetInt("raspistill.camWidth")
	camHeight := viper.GetInt("raspistill.camHeight")
	timeLaps := viper.GetInt("raspistill.timeLaps")
	tmpfsPath := viper.GetString("default.tmpfsPath")
	pictureTempfsCompletPath := path.Join(tmpfsPath, viper.GetString("default.pictureFilename"))
	cmdList := []string{}

	if len(motion) == 0 {
		removeFile()
		log.Critical("Error in config file. motionProgram is not define !")

		os.Exit(1)
	}

	switch method {
	case tmpfs:
		if err := isPrepareTmpfsMethode(); err != nil {
			log.Critical(err)

			return
		}

		cmdList = []string{
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -t 0 -rot %d -tl %d -w %d -h %d -bm", pictureTempfsCompletPath, angle, timeLaps, camWidth, camHeight),
			"mailmotion",
		}

		cmdList = append(cmdList, motion...)

	case motionOnly:
		cmdList = []string{
			"mailmotion",
		}

		cmdList = append(cmdList, motion...)

	case streamerOnly:
		if err := isPrepareTmpfsMethode(); err != nil {
			log.Critical(err)

			return
		}

		cmdList = []string{
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -t 0 -rot %d -tl %d -w %d -h %d -bm", pictureTempfsCompletPath, angle, timeLaps, camWidth, camHeight),
		}

		login := viper.GetString("default.loginCam")
		password := viper.GetString("default.passwordCam")
		port := viper.GetInt("mjpgstreamer.port")
		activateIdentification := viper.GetBool("raspistill.activateIdentification")

		if activateIdentification && len(login) > 0 && len(password) > 0 {
			cmd := fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\" -p %d -c %s:%s", tmpfsPath, "picture.jpg", port, login, password)
			cmdList = append(cmdList, cmd)
		} else {
			cmd := fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\" -p %d", tmpfsPath, "picture.jpg", port)
			cmdList = append(cmdList, cmd)
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

	if err := isAllStartedWithoutGinContext(&cmdList); err != nil {
		killAll(&cmdList)
		removeFile()
	} else {
		gpioStartStop(start)
		// no need createFile() since file already exist !
	}
}

func startAlarm(c *gin.Context) {
	if err := startAlarmWithoutGinContext(); err != nil {
		c.JSON(500, gin.H{"alarm": stop, "stream": stop, "location": viper.GetString("server.location"), "error": err})
	} else {
		c.JSON(200, gin.H{"alarm": start, "stream": stop, "location": viper.GetString("server.location")})
	}
}

func startAlarmWithoutGinContext() error {
	log := logging.MustGetLogger("log")

	method := viper.GetString("default.method")
	motion := viper.GetStringSlice("default.motionProgram")
	angle := viper.GetInt("raspistill.angle")
	camWidth := viper.GetInt("raspistill.camWidth")
	camHeight := viper.GetInt("raspistill.camHeight")
	timeLaps := viper.GetInt("raspistill.timeLaps")
	tmpfsPath := viper.GetString("default.tmpfsPath")
	pictureTempfsCompletPath := path.Join(tmpfsPath, viper.GetString("default.pictureFilename"))
	cmdList := []string{}

	if len(motion) == 0 {
		log.Critical("Error in config file. motionProgram is not define !")

		return errors.New("Error in config file. motionProgram is not define")
	}

	gpioStartStop(start)

	time.Sleep(time.Second * 1)

	switch method {
	case tmpfs:
		if err := isPrepareTmpfsMethode(); err != nil {
			log.Critical(err)
			gpioStartStop(stop)

			return err
		}

		cmdList = []string{
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -t 0 -rot %d -tl %d -w %d -h %d -bm", pictureTempfsCompletPath, angle, timeLaps, camWidth, camHeight),
			"mailmotion",
		}

		cmdList = append(cmdList, motion...)

	case motionOnly:
		cmdList = []string{
			"mailmotion",
		}

		cmdList = append(cmdList, motion...)

	case streamerOnly:
		if err := isPrepareTmpfsMethode(); err != nil {
			log.Critical(err)
			gpioStartStop(stop)

			return err
		}

		cmdList = []string{
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -t 0 -rot %d -tl %d -w %d -h %d -bm", pictureTempfsCompletPath, angle, timeLaps, camWidth, camHeight),
		}

		login := viper.GetString("default.loginCam")
		password := viper.GetString("default.passwordCam")
		port := viper.GetInt("raspistill.port")
		activateIdentification := viper.GetBool("raspistill.activateIdentification")

		if activateIdentification && len(login) > 0 && len(password) > 0 {
			cmd := fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\" -p %d -c %s:%s", tmpfsPath, "picture.jpg", port, login, password)
			cmdList = append(cmdList, cmd)
		} else {
			cmd := fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\" -p %d", tmpfsPath, "picture.jpg", port)
			cmdList = append(cmdList, cmd)
		}

	default:
		errorStr := fmt.Sprintf("Unknown \"%s\" method in config file", method)

		log.Critical(errorStr)
		gpioStartStop(stop)
		removeFile()

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

	// check if all started
	if err := isAllStartedWithoutGinContext(&cmdList); err != nil {
		killAll(&cmdList)
		gpioStartStop(stop)
		removeFile()

		return err
	}

	createFile()
	alarmIsStarted = true

	return nil
}

func stopAlarm(c *gin.Context) {
	stopAlarmWithoutGinContext()

	c.JSON(200, gin.H{"alarm": stop, "stream": stop, "location": viper.GetString("server.location")})
}

func stopAlarmWithoutGinContext() {
	method := viper.GetString("default.method")
	motion := viper.GetStringSlice("default.motionProgram")
	cmdList := []string{}

	switch method {
	case tmpfs:
		cmdList = []string{
			"/opt/vc/bin/raspistill",
			"mailmotion",
		}

		cmdList = append(cmdList, motion...)

	case motionOnly:
		cmdList = []string{
			"mailmotion",
		}

		cmdList = append(cmdList, motion...)

	case streamerOnly:
		cmdList = []string{
			"/opt/vc/bin/raspistill",
			"mjpg_streamer",
		}
	}

	killAll(&cmdList)

	switch method {
	case tmpfs:
		os.Remove(viper.GetString("default.picturePathStore"))
		os.Remove(viper.GetString("default.tmpfsPath"))

	case streamerOnly:
		os.Remove(viper.GetString("default.picturePathStore"))
		os.Remove(viper.GetString("default.tmpfsPath"))
	}

	gpioStartStop(stop)
	removeFile()

	alarmIsStarted = false
}
