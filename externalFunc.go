package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

func createFile() {
	log := logging.MustGetLogger("log")

	slurp, err := os.Create(viper.GetString("default.startalarm"))
	if err != nil {
		log.Criticalf("Unable to create \"%s\" file: %s", viper.GetString("default.startalarm"), err)
	}

	slurp.Close()
}

func getRealExec(cmd *string) {
	*cmd = strings.Replace(*cmd, "LD_LIBRARY_PATH=/usr/lib ", "", -1)
	*cmd = path.Base(strings.Split(*cmd, " ")[0])
}

func isStarted() bool {
	log := logging.MustGetLogger("log")

	if _, err := os.Stat(viper.GetString("default.startalarm")); err != nil {
		log.Debugf("File %s not exist", viper.GetString("default.startalarm"))

		return false
	}

	return true
}

func isInList(progList *[]string, prog *string) bool {
	for _, p := range *progList {
		if p == *prog {
			return true
		}
	}

	return false
}

func killAll(cmdList *[]string) {
	for _, cmd := range *cmdList {
		getRealExec(&cmd)
		cmdStr := fmt.Sprintf("killall -9 %s", cmd)
		exec.Command("/bin/sh", "-c", cmdStr).Output()
	}
}

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

func removeFile() {
	log := logging.MustGetLogger("log")

	if err := os.Remove(viper.GetString("default.startalarm")); err != nil {
		log.Criticalf("Unable to remove \"%s\" file: %s", viper.GetString("default.startalarm"), err)
	}
}
