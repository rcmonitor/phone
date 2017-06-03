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
	"strconv"
)

const(
	_            = iota
	ActionParse  = iota
	ActionCreate = iota
	ActionDrop   = iota
	ActionFormat = iota
	ActionGenerate = iota
	ActionFlush = iota
)

var mapAction = map[string]int{
	"parse":  ActionParse,
	"create":   ActionCreate,
	"drop":   ActionDrop,
	"format": ActionFormat,
	"generate": ActionGenerate,
	"flush": ActionFlush,
}

type TTuple []string

type TApp struct {
	pathOptions *string
	fOptions *os.File
	dbOptions *pg.Options
	db        *pg.DB
	actionString *string
	action    int

	prefix    *string
	codeBegin *string
	codeEnd   *string
	codeSet *string

	f     *os.File
	pathF string

	pathOutput *string
	fOutput *os.File

	errorCounter int
}


func (app *TApp) mProcessArguments() (err error) {

	app.mFlags()
	flag.Parse()

	app.fOptions, err = app.mOpenSourceFile(app.pathOptions)
	if err != nil { return }

	err = app.mReadDBOptions()
	if err != nil { return }

	err = app.mReadAction()
	if err != nil { return }

	app.pathF = flag.Arg(0)

	return
}

func (app *TApp) mReadGenerateFlags() (psSequence *TSequence, err error){
	var intStartCode, intEndCode int
	intStartCode, err = strconv.Atoi(*app.codeBegin)
	if err != nil { return }
	intEndCode, err = strconv.Atoi(*app.codeEnd)
	if err != nil { return }

	var slstrCode []string
	if *app.codeSet != "" {
		slstrCode = strings.Split(*app.codeSet, ",")
	}
	slifCode, err := fStringToIfSlice(slstrCode)
	if err != nil { return }


	if app.f, err = app.mOpenDestinationFile(app.pathF); err != nil { return }

	psSequence = &TSequence{
		Prefix:    *app.prefix,
		StartCode: intStartCode,
		EndCode:   intEndCode,
		SlCode:    slifCode,
		File:      app.f,
	}

	return
}

func (app *TApp) mReadAction() error {
	var ok bool
	if app.action, ok = mapAction[*app.actionString]; !ok {
		return errors.New("Action '" + *app.actionString + "' not implemented")
	}

	return nil
}

func (app *TApp) mOpenSourceFile(pstrFileName *string) (pf *os.File, err error) {
	pf, err = os.Open(*pstrFileName)
	if err != nil {
		err = &TErrorFile{*pstrFileName, err}
	}

	return
}

func (app *TApp) mOpenDestinationFile(strFileName string) (pf *os.File, err error) {
	if strFileName == "" { strFileName = "./data/phone_list.txt" }
	_, err = os.Stat(strFileName)
	if err == nil {
		return os.OpenFile(strFileName, os.O_WRONLY|os.O_APPEND, 0644)
	}

	if !os.IsNotExist(err) { return }

	return os.Create(strFileName)
}

func (app *TApp) mReadDBOptions() error {

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
	DB = pg.Connect(app.dbOptions)
}

func (app *TApp) mRun() error {
	switch app.action {
	case ActionParse:
		return app.mParse()
	case ActionCreate:
		return app.mCreate()
	case ActionDrop:
		return app.mDrop()
	case ActionFlush:
		return app.mFlush()
	case ActionFormat:
		return app.mFormat()
	case ActionGenerate:
		return app.mGenerate()
	default:
		return errors.New("Something went completely wrong")
	}
}

func (app *TApp) mGenerate() (err error) {
	psSequence, err := app.mReadGenerateFlags()
	if err != nil { return }
	defer app.f.Close()

	return psSequence.mGenerate()
}

func (app *TApp) mParse() (err error) {

	if err = app.mOpenSource(); err != nil { return }
	defer app.f.Close()

	csvReader := csv.NewReader(app.f)
	csvReader.Comma = ';'
	csvReader.Comment = '#'
	csvReader.FieldsPerRecord = 6


	var slTuple TTuple

	for{
		slTuple, err = csvReader.Read()
		if err == io.EOF { break }
		if err != nil { return }

		var psdbmPool *TSDBMPool
		psdbmPool, err = NewPool(slTuple)
		if err != nil { return }

		err = psdbmPool.mSave()
		if err != nil { return }
	}

	return nil
}

func (app *TApp) mCreate() error {
	return DB.CreateTable(&TSDBMPool{}, &orm.CreateTableOptions{})
}

func (app *TApp) mDrop() (err error) {
	strQuery := "DROP TABLE IF EXISTS pool"
	_, err = DB.Exec(strQuery)

	return
}

func (app *TApp) mFlush() (err error) {
	strQuery := "UPDATE pool SET generated = false"
	_, err = DB.Exec(strQuery)

	return
}

func (app *TApp) mFormat() (err error) {
	if err = app.mOpenSource(); err != nil { return }
	defer app.f.Close()

	//got no path for output file; let's make it out of input path
	app.mPrepareOutputPath()

	app.fOutput, err = os.Create(*app.pathOutput)
	if err != nil { return }
	defer app.fOutput.Close()

	scSource := bufio.NewScanner(app.f)

	scSource.Scan()
	//let's comment first string which is, kinda, header
	var strFirst string
	if strFirst, err = fWinToUtf(scSource.Text()); err != nil { return err }
	app.fOutput.WriteString("#" + strFirst + "\n")

	replacer := strings.NewReplacer("\t", "", "\"", "\"\"")
	preStrings, err := regexp.Compile(`^(\d+;\d+;\d+;\d+;)([^;]+);([^;]+)$`)
	if err != nil { return }

	var strSource string
	for scSource.Scan() {

		if scSource.Text() == "" { continue }

		//take care of lame encoding
		strSource, err = fWinToUtf(scSource.Text())
		if err != nil { return err }

		//get rid of tabs
		strTemp := replacer.Replace(strSource)
		//add extra quotes
		strNew := preStrings.ReplaceAllString(strTemp, "$1\"$2\";\"$3\"\n")

		if strNew == strTemp {
			rlog.Errorf("no quotes added to: '%s'", strTemp)
			app.errorCounter ++
		}

		if _, err = app.fOutput.WriteString(strNew); err != nil { return }
	}

	err = scSource.Err()
	return
}

func (app *TApp) mShutDown() {
	DB.Close()
	rlog.Debugf("Got %d errors\n", app.errorCounter)
}

func (app *TApp) mPrepareOutputPath() {
	if *app.pathOutput == "" {
		var strOutputFile, strOutputPath string
		strDir, strSourceFile := filepath.Split(app.pathF)
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
	if app.pathF == "" {
		return errors.New("No data file provided")
	}

	var err error

	if app.f, err = app.mOpenSourceFile(&app.pathF); err != nil {
		return &TErrorFile{app.pathF, err}
	}

	return err
}

func (app *TApp) mFlags(){
	app.pathOptions = flag.String("d", "./config/database.json",
		"Path to a json file with database connection information")
	//pfSource := flag.String("s", "./data.csv", "Path to a CSV file with data to process")
	app.actionString = flag.String("a", "parse",
		"Action to perform; \n Available are: \n\t" +
			" - 'parse' parses given csv file; \n\t" +
			" - 'create' generates table in existing database; \n\t" +
			" - 'drop' drops table in existing database; \n\t" +
			" - 'flush' removes 'generated' flag from pools being already used" +
			" - 'format' reformats bogus CSV to real CSV; \n\t" +
			" - 'generate' generates phone numbers for given code range and stores to file")

	app.pathOutput = flag.String("o", "", "Output file path; Used on reformatting.")

	app.prefix = flag.String("p", "", "Prefix to add before each number.")
	app.codeBegin = flag.String("b", "0",
		"Number, beginning of code pool. If omitted, will begin from 0")
	app.codeEnd = flag.String("e", "0",
		"Number, ending of code pool. If omitted, will continue till the maximum code available")
	app.codeSet = flag.String("s", "",
		"Comma-delimeted set of numbers; the only codes to generate; \n" +
			"\tIf provided, -b and -e flags are ignored\n")

	flag.Usage = fUsage
}

