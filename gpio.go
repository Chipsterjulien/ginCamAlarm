package main

import (
	"fmt"
	"os/exec"

	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

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
