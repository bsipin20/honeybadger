package main

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
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

type TransactionProcessor struct {
	store   TransactionStore
	scanner UserInputReader
}

type UserInputReader interface {
	ReadString(prompt string) (string, error)
}

type StdinReader struct {
	scanner *bufio.Scanner
}

func NewStdinReader() *StdinReader {
	return &StdinReader{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (r *StdinReader) ReadString(prompt string) (string, error) {
	fmt.Print(prompt)
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}
		return "", fmt.Errorf("no input received")
	}
	return r.scanner.Text(), nil
}

func NewTransactionProcessor(inputPath string, outputPath string) *TransactionProcessor {
	fm := NewFileManager(inputPath, outputPath)
	if err := fm.Initialize(); err != nil {
		fmt.Printf("Error initializing file manager: %v\n", err)
		os.Exit(1)
	}
	store := NewCSVStore(fm)
	inputReader := NewStdinReader()

	return &TransactionProcessor{
		store:   store,
		scanner: inputReader,
	}
}

func (tp *TransactionProcessor) Run() error {
	rawData := true
	transactions, err := tp.store.ReadTransactions(rawData)
	if err != nil {
		return fmt.Errorf("error reading transactions: %w", err)
	}

	for _, txn := range transactions {
		if err := tp.processTransaction(&txn); err != nil {
			return fmt.Errorf("error processing transaction: %w", err)
		}
	}

	return nil
}

func (tp *TransactionProcessor) processTransaction(txn *Transaction) error {
	fmt.Printf("\nTransaction Details:\n")
	fmt.Printf("Date: %s\n", txn.Date)
	fmt.Printf("Description: %s\n", txn.Description)
	fmt.Printf("Amount: %.2f\n", txn.Amount)

	splitStr, err := tp.scanner.ReadString("Enter split as % of 100 (0 to skip): ")
	if err != nil {
		return fmt.Errorf("error reading split: %w", err)
	}

	split, err := parseSplitPercentage(splitStr)
	if err != nil {
		return fmt.Errorf("invalid split percentage: %w", err)
	}

	if split == 0 {
		txn.Category = CategorySkipped
		txn.Split = 0
		return tp.store.SaveTransaction(txn)
	}

	category, err := tp.getCategory()
	if err != nil {
		return fmt.Errorf("error getting category: %w", err)
	}

	txn.Category = category
	txn.Split = split
	return tp.store.SaveTransaction(txn)
}

func (tp *TransactionProcessor) getCategory() (Category, error) {
	for cat, shortcut := range ValidCategoryShortcuts {
		fmt.Printf("- %s (or '%s')\n", cat, shortcut)
	}

	categoryStr, err := tp.scanner.ReadString("\nEnter category: ")
	if err != nil {
		return "", fmt.Errorf("error reading category: %w", err)
	}

	category, valid := validateCategory(categoryStr)
	if !valid {
		return "", fmt.Errorf("invalid category: %s", categoryStr)
	}

	return category, nil
}

func (tp *TransactionProcessor) createTransaction(record []string) (Transaction, error) {
	floatVal, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		fmt.Println("Error converting string to float64:", err)
		return Transaction{}, err
	}
	return Transaction{
		Date:        record[0],
		Description: record[1],
		Amount:      floatVal,
		SplitID:     uuid.New().String(),
	}, nil
}

const (
	CategoryUtilities     Category = "utilities"
	CategoryGrocery       Category = "grocery"
	CategoryRestaurant    Category = "restaurant"
	CategorySkipped       Category = "skipped"
	CategoryEntertainment Category = "entertainemtn"
	CategoryHome          Category = "home"
)

var ValidCategoryShortcuts = map[Category]string{
	CategoryUtilities:     "u",
	CategoryGrocery:       "g",
	CategoryRestaurant:    "r",
	CategoryEntertainment: "e",
	CategoryHome:          "h",
}

func validateCategory(input string) (Category, bool) {
	input = strings.ToLower(strings.TrimSpace(input))

	for category := range ValidCategoryShortcuts {
		if string(category) == input {
			return category, true
		}
	}

	for category, shortcut := range ValidCategoryShortcuts {
		if shortcut == input {
			return category, true
		}
	}

	return "", false
}

func parseSplitPercentage(input string) (float64, error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return 0, fmt.Errorf("split percentage cannot be empty")
	}

	input = strings.TrimSuffix(input, "%")

	split, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid split percentage format: %w", err)
	}

	if split < 0 || split > 100 {
		return 0, fmt.Errorf("split percentage must be between 0 and 100")
	}

	return split, nil
}
