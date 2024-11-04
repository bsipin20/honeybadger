package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <input.csv> <output.csv>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nExample: %s sept_2024.csv sept_2024_output.csv\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nThe script will:\n")
	fmt.Fprintf(os.Stderr, "1. Read transactions from <input.csv>\n")
	fmt.Fprintf(os.Stderr, "2. Ask for split categories for each transaction\n")
	fmt.Fprintf(os.Stderr, "3. Save results to <output.csv>\n")
	fmt.Fprintf(os.Stderr, "4. Create a progress file at <output.csv>.progress\n")
	os.Exit(1)
}

func main() {
	flag.Usage = usage

	flag.Parse()

	if flag.NArg() != 2 {
		usage()
	}

	inputFile := flag.Arg(0)
	outputFile := flag.Arg(1)

	fm := NewFileManager(inputFile, outputFile)

	if err := fm.Initialize(); err != nil {
		fmt.Printf("Error initializing file manager: %v\n", err)
		os.Exit(1)
	}

	processor := NewTransactionProcessor(fm)
	if err := processor.ProcessTransactions(); err != nil {
		fmt.Printf("Error processing transactions: %v\n", err)
		os.Exit(1)
	}
}
