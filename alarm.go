package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/gin-gonic/gin"
	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

func isAllStarted(c *gin.Context, cmdList *[]string) error {
	log := logging.MustGetLogger("log")

	for _, cmd := range *cmdList {
		// cmd = strings.Replace(cmd, "LD_LIBRARY_PATH=/usr/lib ", "", -1)
		// realExec := path.Base(strings.Split(cmd, " ")[0])
		// cmdStr := fmt.Sprintf("pgrep ^%s$", realExec)
		getRealExec(&cmd)

		cmdStr := fmt.Sprintf("pgrep ^%s$", cmd)

		log.Debugf("Cmd to exec: %s", cmdStr)

		if _, err := exec.Command("/bin/sh", "-c", cmdStr).Output(); err != nil {
			// errorStr := fmt.Sprintf("\"%s\" is not running !", realExec)
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
		// cmd = strings.Replace(cmd, "LD_LIBRARY_PATH=/usr/lib ", "", -1)
		// realExec := path.Base(strings.Split(cmd, " ")[0])
		// cmdStr := fmt.Sprintf("pgrep ^%s$", realExec)
		getRealExec(&cmd)
		cmdStr := fmt.Sprintf("pgrep ^%s$", cmd)

		log.Debugf("Cmd to exec: %s", cmdStr)

		if _, err := exec.Command("/bin/sh", "-c", cmdStr).Output(); err != nil {
			// errorStr := fmt.Sprintf("\"%s\" is not running !", realExec)
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

// func isPrepareTmpfsMethodeWithoutGinContext() bool {
// 	log := logging.MustGetLogger("log")
//
// 	tmpfs := viper.GetString("default.tmpfsPath")
// 	pictureTempfsCompletPath := path.Join(tmpfs, viper.GetString("default.pictureFilename"))
// 	pictureStoreCompletPath := path.Join(viper.GetString("default.picturePathStore"), viper.GetString("default.dirSendName"))
//
// 	log.Debug(pictureTempfsCompletPath)
// 	log.Debug(pictureStoreCompletPath)
//
// 	// check if picture file exist
// 	if _, err := os.Stat(pictureTempfsCompletPath); err == nil {
// 		if er := os.Remove(pictureTempfsCompletPath); er != nil {
// 			errorStr := fmt.Sprintf("Unable to remove \"%s\" in tmpfs method: %s", pictureTempfsCompletPath, err)
//
// 			log.Criticalf(errorStr)
//
// 			return false
// 		}
// 	} else {
// 		// try to create a file to test write right
// 		randFilename := randStringBytesMaskImprSrc(10)
// 		randFilenameFullPath := path.Join(tmpfs, randFilename)
//
// 		fn, err := os.Create(path.Join(tmpfs, randFilename))
// 		if err != nil {
// 			errorStr := fmt.Sprintf("Unable to create a random file in \"%s\": %s", randFilenameFullPath, err)
//
// 			log.Critical(errorStr)
//
// 			return false
// 		}
//
// 		fn.Close()
//
// 		if err = os.Remove(randFilenameFullPath); err != nil {
// 			errorStr := fmt.Sprintf("Unable to remove \"%s\": %s", randFilenameFullPath, err)
//
// 			log.Critical(errorStr)
//
// 			return false
// 		}
//
// 		if err := os.MkdirAll(pictureStoreCompletPath, 0744); err != nil {
// 			errorStr := fmt.Sprintf("Unable to create \"%s\" directories: %s", pictureStoreCompletPath, err)
//
// 			log.Critical(errorStr)
//
// 			return false
// 		}
// 	}
//
// 	return true
// }

func randStringBytesMaskImprSrc(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, srcRand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = srcRand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func restartAlarm() {
	log := logging.MustGetLogger("log")

	method := viper.GetString("default.method")
	// motion := viper.GetString("default.motionProgram")
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
			// motion,
			"mailmotion",
		}

		cmdList = append(cmdList, motion...)

	case motionOnly:
		cmdList = []string{
			// motion,
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
			fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\"", tmpfsPath, "picture.jpg"),
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
	log := logging.MustGetLogger("log")

	method := viper.GetString("default.method")
	// motion := viper.GetString("default.motionProgram")
	motion := viper.GetStringSlice("default.motionProgram")
	angle := viper.GetInt("raspistill.angle")
	camWidth := viper.GetInt("raspistill.camWidth")
	camHeight := viper.GetInt("raspistill.camHeight")
	timeLaps := viper.GetInt("raspistill.timeLaps")
	tmpfsPath := viper.GetString("default.tmpfsPath")
	pictureTempfsCompletPath := path.Join(tmpfsPath, viper.GetString("default.pictureFilename"))
	cmdList := []string{}

	// if motion == "" {
	if len(motion) == 0 {
		log.Critical("Error in config file. motionProgram is not define !")
		c.JSON(500, gin.H{"error": "Error in config file. motionProgram is not define !"})

		return
	}

	switch method {
	case tmpfs:
		if err := isPrepareTmpfsMethode(); err != nil {
			log.Critical(err)
			c.JSON(500, gin.H{"error": err})

			return
		}

		cmdList = []string{
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -t 0 -rot %d -tl %d -w %d -h %d -bm", pictureTempfsCompletPath, angle, timeLaps, camWidth, camHeight),
			// motion,
			"mailmotion",
		}

		cmdList = append(cmdList, motion...)

	case motionOnly:
		cmdList = []string{
			// motion,
			"mailmotion",
		}

		cmdList = append(cmdList, motion...)

	case streamerOnly:
		if err := isPrepareTmpfsMethode(); err != nil {
			log.Critical(err)
			c.JSON(500, gin.H{"error": err})

			return
		}

		cmdList = []string{
			fmt.Sprintf("/opt/vc/bin/raspistill -o %s -t 0 -rot %d -tl %d -w %d -h %d -bm", pictureTempfsCompletPath, angle, timeLaps, camWidth, camHeight),
			fmt.Sprintf("LD_LIBRARY_PATH=/usr/lib mjpg_streamer -b -i \"input_file.so -f %s -n %s\" -o \"output_http.so -w /usr/share/mjpg-streamer/www/\"", tmpfsPath, "picture.jpg"),
		}

	default:
		errorStr := fmt.Sprintf("Unknown \"%s\" method in config file", method)

		log.Critical(errorStr)
		c.JSON(500, gin.H{"error": errorStr})

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
	if err := isAllStarted(c, &cmdList); err != nil {
		killAll(&cmdList)
		removeFile()

		c.JSON(500, gin.H{"alarm": stop, "stream": stop, "location": viper.GetString("server.location")})
	} else {
		gpioStartStop(start)
		createFile()

		c.JSON(200, gin.H{"alarm": start, "stream": stop, "location": viper.GetString("server.location")})
	}
}

func stopAlarm(c *gin.Context) {
	method := viper.GetString("default.method")
	// motion := viper.GetString("default.motionProgram")
	motion := viper.GetStringSlice("default.motionProgram")
	cmdList := []string{}

	switch method {
	case tmpfs:
		cmdList = []string{
			"/opt/vc/bin/raspistill",
			// motion,
			"mailmotion",
		}

		cmdList = append(cmdList, motion...)

	case motionOnly:
		cmdList = []string{
			// motion,
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

	c.JSON(200, gin.H{"alarm": stop, "stream": stop, "location": viper.GetString("server.location")})
}

func killAll(cmdList *[]string) {
	for _, cmd := range *cmdList {
		// realExec := path.Base(strings.Split(cmd, " ")[0])
		// cmdStr := fmt.Sprintf("killall -9 %s", realExec)
		// exec.Command("/bin/sh", "-c", cmdStr).Output()
		getRealExec(&cmd)
		cmdStr := fmt.Sprintf("killall -9 %s", cmd)
		exec.Command("/bin/sh", "-c", cmdStr).Output()
	}
}
