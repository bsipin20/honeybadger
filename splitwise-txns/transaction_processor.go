package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	"io"
	"os"
	"strconv"
	"strings"
)

type Category string

type CategoryOption struct {
	Full     Category
	Shortcut string
}

var categoryOptions = []CategoryOption{
	{Full: "utilities", Shortcut: "u"},
	{Full: "home", Shortcut: "h"},
	{Full: "health", Shortcut: "h"},
	{Full: "restaurant", Shortcut: "r"},
	{Full: "entertainment", Shortcut: "e"},
}

func isValidCategory(input string) (Category, bool) {
	inputLower := strings.ToLower(strings.TrimSpace(input))

	for _, cat := range categoryOptions {
		if inputLower == string(cat.Full) || inputLower == cat.Shortcut {
			return cat.Full, true
		}
	}
	return "", false
}

type Transaction struct {
	Date        string
	Description string
	Amount      float64
	Split       string
	Category    Category
	SplitID     string
}

type TransactionProcessor struct {
	fileManager *FileManager
	scanner     *bufio.Scanner
}

func NewTransactionProcessor(fm *FileManager) *TransactionProcessor {
	return &TransactionProcessor{
		fileManager: fm,
		scanner:     bufio.NewScanner(os.Stdin),
	}
}

func (tp *TransactionProcessor) ProcessTransactions() error {
	file, err := os.Open(tp.fileManager.InputPath)
	if err != nil {
		return fmt.Errorf("error opening input file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("error reading header: %w", err)
	}

	for i := 0; i < tp.fileManager.LastPosition; i++ {
		if _, err := reader.Read(); err != nil {
			return fmt.Errorf("error skipping to last position: %w", err)
		}
	}

	return tp.processRemainingTransactions(reader)
}

func (tp *TransactionProcessor) processRemainingTransactions(reader *csv.Reader) error {
	lineNum := tp.fileManager.LastPosition

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading record: %w", err)
		}

		transaction, err := tp.createTransaction(record)

		if err := tp.displayAndProcessTransaction(&transaction, lineNum); err != nil {
			return err
		}

		if err := tp.saveTransaction(transaction); err != nil {
			return err
		}

		lineNum++
		if err := tp.fileManager.saveProgress(lineNum); err != nil {
			return fmt.Errorf("error saving progress: %w", err)
		}
	}

	fmt.Println("\nAll transactions processed!")
	return nil
}

func (tp *TransactionProcessor) createTransaction(record []string) (Transaction, error) {
  floatVal, err := strconv.ParseFloat(record[2], 64)
  if err != nil {
        fmt.Println("Error converting string to float64:", err)
        return Transaction{} , err
  }
	return Transaction{
		Date:        record[0],
		Description: record[1],
		Amount:      floatVal,
		SplitID:     uuid.New().String(),
	}, nil
}

func (tp *TransactionProcessor) displayAndProcessTransaction(t *Transaction, lineNum int) error {

	fmt.Printf("\nTransaction #%d:\n", lineNum+1)
	fmt.Printf("Date: %s\n", t.Date)
	fmt.Printf("Description: %s\n", t.Description)
	fmt.Printf("Amount: %s\n", t.Amount)

	fmt.Print("Enter split as % of 100: ")

	if !tp.scanner.Scan() {
		return fmt.Errorf("error reading split input")
	}

	splitStr := strings.TrimSpace(tp.scanner.Text())
	split, err := strconv.ParseFloat(splitStr, 64)
	if err != nil {
		return fmt.Errorf("invalid split percentage: %v", err)
	}

	if split == 0 {
		t.Split = "skipped"
		fmt.Println("Transaction skipped")
		return nil
	} else {
		t.Split = splitStr
	}

	fmt.Println("\nAvailable categories:")
	for _, cat := range categoryOptions {
		fmt.Printf("- %s (or '%s')\n", cat.Full, cat.Shortcut)
	}

	for {
		fmt.Print("\nEnter category: ")

		if !tp.scanner.Scan() {
			return fmt.Errorf("error reading user input")
		}

		input := strings.TrimSpace(tp.scanner.Text())
		if input == "" {
			fmt.Println("Category cannot be empty. Please try again.")
			continue
		}

		if category, valid := isValidCategory(input); valid {
			t.Category = category
			fmt.Printf("Recorded category as: %s\n", t.Split)
			return nil
		}

		fmt.Println("Invalid category. Please choose from the available options.")
		continue
	}

	return nil
}

func (tp *TransactionProcessor) saveTransaction(t Transaction) error {
	outFile, err := os.OpenFile(tp.fileManager.OutputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening output file: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	if err := writer.Write([]string{
		t.Date,
		t.Description,
    fmt.Sprintf("%.2f", t.Amount),
		t.Split,
    fmt.Sprint(t.Category),         // Convert to string if needed
		t.SplitID,
	}); err != nil {
		return fmt.Errorf("error writing transaction: %w", err)
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing writer: %w", err)
	}

	fmt.Printf("Split recorded with ID: %s\n", t.SplitID)
	return nil
}
