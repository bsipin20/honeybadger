package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ProjectName string
	InputFile   string
	Process     bool
	Summarize   bool
}

func parseFlags() Config {
	var config Config

	flag.BoolVar(&config.Process, "process", false, "Process transactions")
	flag.BoolVar(&config.Summarize, "summarize", false, "Summarize transactions")
	flag.StringVar(&config.ProjectName, "project-name", "", "Project name (required)")
	flag.StringVar(&config.InputFile, "input", "", "Input CSV file (required for process)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of transaction processor:\n\n")
		fmt.Fprintf(os.Stderr, "Process transactions:\n")
		fmt.Fprintf(os.Stderr, "  txn --process --project-name=feb_2024 --input=feb_2024.csv\n\n")
		fmt.Fprintf(os.Stderr, "Summarize transactions:\n")
		fmt.Fprintf(os.Stderr, "  txn --summarize --project-name=feb_2024\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if err := validateFlags(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	return config
}

func validateFlags(config Config) error {
	if config.ProjectName == "" {
		return fmt.Errorf("project-name is required")
	}

	if !config.Process && !config.Summarize {
		return fmt.Errorf("either --process or --summarize must be specified")
	}

	if config.Process && config.Summarize {
		return fmt.Errorf("cannot specify both --process and --summarize")
	}

	if config.Process && config.InputFile == "" {
		return fmt.Errorf("--input is required when using --process")
	}

	return nil
}

func handleProcess(inputPath string, outputPath string) error {
	//	fm := NewFileManager(inputPath, outputPath)

	//	csvStore := NewCSVStore(fm)

	proc := NewTransactionProcessor(inputPath, outputPath)

	fmt.Printf("Processing transactions from %s\n", inputPath)
	fmt.Printf("Saving results to %s\n", outputPath)

	if err := proc.Run(); err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	fmt.Println("Processing complete!")
	return nil
}

func handleSummarize(transactionsPath string) error {
	sum := NewSummarizeProcessor(transactionsPath)

	fmt.Printf("Generating summary for %s\n", transactionsPath)

	if err := sum.Run([]string{transactionsPath}); err != nil {
		return fmt.Errorf("summarizing failed: %w", err)
	}

	return nil
}

func run(config Config) error {
	projectDir := filepath.Join("data", config.ProjectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	outputPath := filepath.Join(projectDir, "transactions.csv")

	if config.Process {
		return handleProcess(config.InputFile, outputPath)
	}
	return handleSummarize(outputPath)
}

func main() {
	config := parseFlags()

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
