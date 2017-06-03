package main

import (
	"github.com/romana/rlog"
	"os"
	"github.com/go-pg/pg"
	"time"
)

const (
	ExitOk = iota
	ExitError = iota
)

func init() {
	rlog.SetConfFile("config/log.conf")
}

var DB *pg.DB

func main() {
	var intExitStatus int = ExitOk

	tStart := time.Now()
	rlog.Infof("Started at: %s", tStart.String())

	psApp := &TApp{}
	err := psApp.mProcessArguments()
	if err != nil {
		rlog.Error(err)
		intExitStatus = ExitError
	}else{
		psApp.mInit()

		err = psApp.mRun()
		if err != nil {
			rlog.Error(err)
			intExitStatus = ExitError
		}

		psApp.mShutDown()
		tEnd := time.Now()
		tiPassed := tEnd.Sub(tStart)
		rlog.Infof("Execution took %s", tiPassed.String())
	}

	os.Exit(intExitStatus)
}

