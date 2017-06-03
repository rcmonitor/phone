package main

import (
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/charmap"
	"strings"
	"io/ioutil"
	"strconv"
	"github.com/romana/rlog"
)

func fWinToUtf(strSource string) (string, error) {
	srSource := strings.NewReader(strSource)

	tr := transform.NewReader(srSource, charmap.Windows1251.NewDecoder())
	buf, err := ioutil.ReadAll(tr)
	if err != nil { return "", err }

	return string(buf), nil
}

func fStringToIfSlice(slstr []string) (slif []interface{}, err error) {
	var intTemp int
	for _, str := range slstr {
		rlog.Debugf("Got string from code slice: '%s'", str)
		if intTemp, err = strconv.Atoi(str); err != nil { return }
		slif = append(slif, intTemp)
	}
	rlog.Debug("No code provided in slice")

	return
}