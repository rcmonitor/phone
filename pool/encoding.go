package main

import (
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/charmap"
	"strings"
	"io/ioutil"
)

func fWinToUtf(strSource string) (string, error) {
	srSource := strings.NewReader(strSource)

	tr := transform.NewReader(srSource, charmap.Windows1251.NewDecoder())
	buf, err := ioutil.ReadAll(tr)
	if err != nil { return "", err }

	return string(buf), nil
}