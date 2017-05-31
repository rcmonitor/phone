package main

import (
	"github.com/go-pg/pg"
	"strconv"
)


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

func (psPool *TSDBMPool) mSave(db *pg.DB) error {
	return db.Insert(psPool)
}

