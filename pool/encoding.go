package main

import (
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/charmap"
	"strings"
	"io/ioutil"
	"strconv"
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
		if intTemp, err = strconv.Atoi(str); err != nil { return }
		slif = append(slif, intTemp)
	}

	return
}