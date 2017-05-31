package main

import "fmt"

type TErrorFile struct {
	FileName string
	Err error
}

func (psError *TErrorFile) Error() string {
	return fmt.Sprintf("'%s': '%s'", psError.FileName, psError.Err.Error())
}
