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
    var summarizeFlag bool
    flag.BoolVar(&summarizeFlag, "summarize", false, "Summarize existing transaction file")
    
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage:\n")
        fmt.Fprintf(os.Stderr, "  %s <input.csv> <output.csv>    # Process transactions\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "  %s --summarize <output.csv>    # Summarize existing transactions\n\n", os.Args[0])
        flag.PrintDefaults()
    }

    flag.Parse()

    if summarizeFlag {
        if flag.NArg() != 1 {
            fmt.Println("Error: Please provide the transaction file to summarize")
            flag.Usage()
            os.Exit(1)
        }

        if err := handleSummarize(flag.Arg(0)); err != nil {
            fmt.Printf("Error summarizing transactions: %v\n", err)
            os.Exit(1)
        }
        return

    } else {

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
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	factory := NewProcessorFactory()
	processor, err := factory.GetProcessor(command)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		usage()
		os.Exit(1)
	}

	if err := processor.Run(args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}


