package main

import (
	"strconv"
	"github.com/go-pg/pg"
	"io"
)

type TSLPool []*TSDBMPool

type TSDBMPool struct {
	tableName struct{} `sql:"pool"`
	Id int `sql:",pk"`
	Code int
	ValueStart int
	ValueEnd int
	Capacity int
	Provider string
	Region string
	Generated bool `sql:",notnull"`
}

func NewPool(slTuple TTuple) (*TSDBMPool, error) {
	var err error
	var intCode, intValueStart, intValueEnd, intCapacity int
	intCode, err = strconv.Atoi(slTuple[0])
	if err != nil { return nil, err }

	intValueStart, err = strconv.Atoi(slTuple[1])
	if err != nil { return nil, err }

	intValueEnd, err = strconv.Atoi(slTuple[2])
	if err != nil { return nil, err }

	intCapacity, err = strconv.Atoi(slTuple[3])
	if err != nil { return nil, err }

	return &TSDBMPool{
		Code: intCode,
		ValueStart: intValueStart,
		ValueEnd: intValueEnd,
		Capacity: intCapacity,
		Provider: slTuple[4],
		Region: slTuple[5],
	}, nil
}

func (psPool *TSDBMPool) mSave() error {
	return DB.Insert(psPool)
}

func (psPool *TSDBMPool) mGenerate(pfSequence io.Writer, strPrefix string) (err error) {
	for i := psPool.ValueStart; i <= psPool.ValueEnd; i ++ {
		strPhone := strPrefix + strconv.Itoa(psPool.Code) + strconv.Itoa(i) + "\n"
		_, err = pfSequence.Write([]byte(strPhone))
		if err != nil { return err }
	}

	psPool.Generated = true
	return psPool.mUpdateGenerated()
}

func (psPool *TSDBMPool) mUpdateGenerated() (err error) {
	_, err = DB.Model(psPool).Column("generated").Update()
	return
}

func fGetMaxCode() (int, error) {
	var intMaxCode int
	_, err := DB.QueryOne(pg.Scan(&intMaxCode), "SELECT MAX(code) FROM pool")

	return intMaxCode, err
}

