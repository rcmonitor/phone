package main

import (
	"os"
	"github.com/romana/rlog"
)

type TSequence struct {
	Prefix string
	StartCode int
	EndCode int
	EndCodeVerified int
	SlCode []interface{}
	File *os.File

	pool TSLPool
	limit int
	page int
	lastProcessedCode int
}

func (psSequence *TSequence) mGenerate() (err error) {

	psSequence.limit = 1000

	if psSequence.EndCode == 0 {
		psSequence.EndCodeVerified, err = fGetMaxCode()
		if err != nil { return }
	}else{ psSequence.EndCodeVerified = psSequence.EndCode }

	psSequence.mLogCodes()

	defer rlog.Infof("Finished on code %d", psSequence.lastProcessedCode)

	for {
		err = psSequence.mGetNextPool()
		if err != nil { return }

		for _, psPool := range psSequence.pool {
			err = psPool.mGenerate(psSequence.File, psSequence.Prefix)
			if err != nil { return }
			psSequence.lastProcessedCode = psPool.Code
		}

		intLength := len(psSequence.pool)
		//this was the last page
		if intLength < psSequence.limit || intLength == 0 {
			break
		}
		psSequence.page ++
	}

	return err
}

func (psSequence *TSequence) mGetNextPool() (err error) {

	if psSequence.mIsCodeSet() {
		err = DB.Model(&psSequence.pool).
			WhereIn("tsdbm_pool.code IN (?)", psSequence.SlCode...).
			Where("tsdbm_pool.generated = ?", false).
			Order("code ASC").
			Order("value_start ASC").
			Limit(psSequence.limit).
			Offset(psSequence.page * psSequence.limit).
			Select()
	}else{
		err = DB.Model(&psSequence.pool).
			Where("tsdbm_pool.code >= ?", psSequence.StartCode).
			Where("tsdbm_pool.code <= ?", psSequence.EndCodeVerified).
			Where("tsdbm_pool.generated = ?", false).
			Order("code ASC").
			Order("value_start ASC").
			Limit(psSequence.limit).
			Offset(psSequence.page * psSequence.limit).
			Select()
	}

	return
}

func (psSequence *TSequence) mLogCodes() {
	if psSequence.mIsCodeSet() {
		rlog.Infof("Going to generate codes: %v", psSequence.SlCode)
	}else{
		rlog.Infof("Going to generate codes %d through %d",
			psSequence.StartCode, psSequence.EndCodeVerified)
	}
}

func (psSequence *TSequence) mIsCodeSet() bool {
	return len(psSequence.SlCode) > 0
}
