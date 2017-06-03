package main

import (
	"fmt"
	"flag"
)

func fUsage() {
	fmt.Println("Tool to generate phone list out of CSV-formatted pool data")
	fmt.Println("Generation may be continued from last position if somehow interrupted")
	fmt.Println("Safe to specify same code for same file many times: it will be skipped if not flushed")
	fmt.Println("\nUsage:")
	fmt.Println("1st form: pool -a (create|drop|flush) [-d <path to json db settings file>]")
	fmt.Println("2nd form: pool -a format [-o <path to utf-8 CSV output file>] <path to CSV input file>")
	fmt.Println("3rd form: pool -a parse [-d <path to json db settings file>] <path to CSV input file>")
	fmt.Println("4th form: pool -a generate [-p <prefix>] [-s <start code>] [-e <end code>] [<path to phone list output file>]")
	fmt.Println("path to rlog config file: './config/log.conf'")
	fmt.Println("\nDefaults:")
	fmt.Println("<path to json db settings file>: './config/database.json'")
	fmt.Println("<path to phone list output file>: './data/phone_list.txt'")
	fmt.Println("<path to utf-8 CSV output file>: '<path to CSV input file>_formatted.csv'\n")
	flag.PrintDefaults()
}

