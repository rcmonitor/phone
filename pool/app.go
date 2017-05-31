package main

import (
	"flag"
	"github.com/go-pg/pg"
	"os"
	"io/ioutil"
	"encoding/json"
	"errors"
	"github.com/go-pg/pg/orm"
	"github.com/romana/rlog"
	"encoding/csv"
	"io"
	"path/filepath"
	"strings"
	"bufio"
	"regexp"
	"fmt"
)

const(
	_ = iota
	ActionParse = iota
	ActionGenerate = iota
	ActionDrop = iota
	ActionFormat = iota
)

var mapAction = map[string]int{
	"parse": ActionParse,
	"generate": ActionGenerate,
	"drop": ActionDrop,
	"format": ActionFormat,
}

type TTuple []string

type TApp struct {
	pathOptions *string
	fOptions *os.File
	dbOptions *pg.Options
	db        *pg.DB
	action    int

	fCSV *os.File
	pathCSV string

	pathOutput *string
	fOutput *os.File
}


func (app *TApp) mProcessArguments() (err error) {

	app.pathOptions = flag.String("d", "./config/database.json",
		"Path to a json file with database connection information")
	//pfSource := flag.String("s", "./data.csv", "Path to a CSV file with data to process")
	pstrAction := flag.String("a", "parse",
		"Action to perform; \n Available are: \n\t" +
			" - 'parse' parses given csv file; \n\t" +
			" - 'generate' generates table in existing database; \n\t" +
			" - 'drop' drops table in existing database; \n\t" +
			" - 'format' reformats bogus CSV to real CSV")

	app.pathOutput = flag.String("o", "", "Output file path; Used on reformatting.")

	flag.Usage = fUsage
	flag.Parse()

	app.fOptions, err = app.mOpenFile(app.pathOptions)
	if err != nil { return }

	err = app.mReadOptions()
	if err != nil {
		rlog.Debug(err)
		return }

	err = app.mReadAction(pstrAction)
	if err != nil { return }

	app.pathCSV = flag.Arg(0)

	return
}

func (app *TApp) mReadAction(pstrAction *string) error {
	var ok bool
	if app.action, ok = mapAction[*pstrAction]; !ok {
		return errors.New("Action '" + *pstrAction + "' not implemented")
	}

	return nil
}

func (app *TApp) mOpenFile(pstrFileName *string) (pf *os.File, err error) {
	pf, err = os.Open(*pstrFileName)
	if err != nil {
		err = &TErrorFile{*pstrFileName, err}
	}

	return
}

func (app *TApp) mReadOptions() error {

	defer app.fOptions.Close()

	slbDBOptions, err := ioutil.ReadAll(app.fOptions)
	if err != nil {
		return &TErrorFile{*app.pathOptions, err}
	}

	app.dbOptions = &pg.Options{}
	err = json.Unmarshal(slbDBOptions, app.dbOptions)
	if err != nil {
		return &TErrorFile{*app.pathOptions, err}
	}

	return nil
}

func (app *TApp) mInit(){
	app.db = pg.Connect(app.dbOptions)
}

func (app *TApp) mRun() error {
	switch app.action {
	case ActionParse:
		return app.mParse()
	case ActionGenerate:
		return app.mGenerate()
	case ActionDrop:
		return app.mDrop()
	case ActionFormat:
		return app.mFormat()
	default:
		return errors.New("Something went completely wrong")
	}
}

func (app *TApp) mParse() (err error) {

	if err = app.mOpenSource(); err != nil { return }
	defer app.fCSV.Close()

	csvReader := csv.NewReader(app.fCSV)
	csvReader.Comma = ';'
	csvReader.Comment = '#'
	csvReader.FieldsPerRecord = 6


	var slTuple TTuple
	//slTuple, err = csvReader.Read()
	//if err != nil { return }
	//rlog.Infof("CSV header: '%v'", slTuple)

	for{
		slTuple, err = csvReader.Read()
		if err == io.EOF { break }
		if err != nil { return }

		var psdbmPool *TSDBMPool
		psdbmPool, err = NewPool(slTuple)
		if err != nil { return }

		err = psdbmPool.mSave(app.db)
		if err != nil { return }
	}

	return nil
}

func (app *TApp) mGenerate() error {
	return app.db.CreateTable(&TSDBMPool{}, &orm.CreateTableOptions{})
}

func (app *TApp) mDrop() (err error) {
	strQuery := "DROP TABLE IF EXISTS pool"
	_, err = app.db.Exec(strQuery)

	return
}

func (app *TApp) mFormat() (err error) {
	if err = app.mOpenSource(); err != nil { return }
	defer app.fCSV.Close()

	//got no path for output file; let's make it out of input path
	app.mPrepareOutputPath()

	app.fOutput, err = os.Create(*app.pathOutput)
	if err != nil { return }
	defer app.fOutput.Close()

	scSource := bufio.NewScanner(app.fCSV)

	scSource.Scan()
	//let's comment first string which is, kinda, header
	var strFirst string
	if strFirst, err = fWinToUtf(scSource.Text()); err != nil { return err }
	app.fOutput.WriteString("#" + strFirst + "\n")

	replacer := strings.NewReplacer("\t", "", "\"", "\"\"")
	preStrings, err := regexp.Compile(`^(\d+;\d+;\d+;\d+;)(\D+);(\D+)$`)
	if err != nil { return }

	var strSource string
	for scSource.Scan() {

		//take care of lame encoding
		strSource, err = fWinToUtf(scSource.Text())
		if err != nil { return err }

		strNew := replacer.Replace(strSource)
		strNew = preStrings.ReplaceAllString(strNew, "$1\"$2\";\"$3\"\n")

		if _, err = app.fOutput.WriteString(strNew); err != nil { return }
	}

	err = scSource.Err()

	return
}

func (app *TApp) mShutDown() {
	app.db.Close()
}

func (app *TApp) mPrepareOutputPath() {
	if *app.pathOutput == "" {
		var strOutputFile, strOutputPath string
		strDir, strSourceFile := filepath.Split(app.pathCSV)
		strExt := filepath.Ext(strSourceFile)
		if strExt != "" {
			strOutputFile = strings.Trim(strSourceFile, "." + strExt)
		}else{
			strOutputFile = strSourceFile
		}

		strOutputPath = strDir + strOutputFile + "_formatted.csv"
		app.pathOutput = &strOutputPath
	}
}

func (app *TApp) mOpenSource() error {
	if app.pathCSV == "" {
		return errors.New("No data file provided")
	}

	var err error

	if app.fCSV, err = app.mOpenFile(&app.pathCSV); err != nil {
		return &TErrorFile{app.pathCSV, err}
	}

	return err
}

func fUsage() {
	fmt.Println("Usage:")
	fmt.Println("1st form: pool -a (generate|drop) [-d <path to db settings>]")
	fmt.Println("2nd form: pool -a format [-o <path to output file>] <path to input file>")
	fmt.Println("3rd form: pool -a parse [-d <path to db settings>] <path to input file>")
	flag.PrintDefaults()
}
