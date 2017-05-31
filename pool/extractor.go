package main

import (
	"github.com/romana/rlog"
	"os"
)

const (
	ExitOk = iota
	ExitError = iota
)

func init() {
	rlog.SetConfFile("config/log.conf")
}

func main() {
	var intExitStatus int = ExitOk

	psApp := &TApp{}
	err := psApp.mProcessArguments()
	if err != nil {
		rlog.Error(err)
		os.Exit(1)
	}

	psApp.mInit()

	err = psApp.mRun()
	if err != nil {
		rlog.Error(err)
		intExitStatus = ExitError
	}

	psApp.mShutDown()
	os.Exit(intExitStatus)
}

