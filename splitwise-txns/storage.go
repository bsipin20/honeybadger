package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type HeaderConfig struct {
	Names    []string
	DatePos  int
	DescPos  int
	AmtPos   int
	CatPos   int
	SplitPos int
}

var DefaultHeaders = HeaderConfig{
	Names:    []string{"Date", "Description", "Amount", "Category", "Split"},
	DatePos:  0,
	DescPos:  1,
	AmtPos:   2,
	CatPos:   3,
	SplitPos: 4,
}

type TransactionStore interface {
	ReadTransactions() ([]Transaction, error)
	SaveTransaction(transaction Transaction) error
	//	UpdateTransaction(filename string, id string, transaction model.Transaction)
}

type Transaction struct {
	Date        string
	Description string
	Amount      float64
	Split       float64
	Category    Category
	SplitID     string
}

type CSVStore struct {
	fileManager *FileManager
	headers     HeaderConfig
}

// TODO implement option to write files to directory based on project name
func NewCSVStore(fm *FileManager) *CSVStore {
	return &CSVStore{
		fileManager: fm,
		headers:     DefaultHeaders,
	}
}

func (s *CSVStore) ReadTransactions() ([]Transaction, error) {
	file, err := os.Open(s.fileManager.InputPath)

	if err != nil {
		return nil, fmt.Errorf("error opening new ile %w", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()

	if err != nil {
		return nil, fmt.Errorf("error reading header: %w", err)
	}

	if err := s.validateHeader(header); err != nil {
		return nil, err
	}

	var transactions []Transaction

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("error reading record: %w", err)
		}

		transaction, err := s.parseTransaction(record)

		if err != nil {
			return nil, fmt.Errorf("error parsing record: %w", err)
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (s *CSVStore) createOutputFileWithHeaders() error {
    file, err := os.Create(s.fileManager.OutputPath)
    if err != nil {
        return fmt.Errorf("error creating file: %w", err)
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    headers := []string{"Date", "Description", "Amount", "Category", "Split"}
    if err := writer.Write(headers); err != nil {
        return fmt.Errorf("error writing headers: %w", err)
    }

    return nil
}


func (fm *FileManager) GetCurrentPosition() (int, error) {
    if _, err := os.Stat(fm.ProgressPath); os.IsNotExist(err) {
        return 0, nil
    }

    content, err := os.ReadFile(fm.ProgressPath)
    if err != nil {
        return 0, fmt.Errorf("error reading progress file: %w", err)
    }

    var position int
    if _, err := fmt.Sscanf(string(content), "%d", &position); err != nil {
        return 0, fmt.Errorf("error parsing progress value: %w", err)
    }

    return position, nil
}

func (fm *FileManager) SaveProgress(position int) error {
    dir := filepath.Dir(fm.ProgressPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("error creating progress directory: %w", err)
    }

    content := fmt.Sprintf("%d", position)
    if err := os.WriteFile(fm.ProgressPath, []byte(content), 0644); err != nil {
        return fmt.Errorf("error writing progress file: %w", err)
    }

    fm.LastPosition = position
    return nil
}

func (s *CSVStore) SaveTransaction(transaction Transaction) error {
	dir := filepath.Dir(s.fileManager.OutputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dir, err)
	}

	if _, err := os.Stat(s.fileManager.OutputPath); os.IsNotExist(err) {
		if err := s.createOutputFileWithHeaders(); err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
	}

	file, err := os.OpenFile(s.fileManager.OutputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file for append: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{
		transaction.Date,
		transaction.Description,
		fmt.Sprintf("%.2f", transaction.Amount),
		string(transaction.Category),
		fmt.Sprintf("%.2f", transaction.Split),
	}

	if err := writer.Write(record); err != nil {
		return fmt.Errorf("error writing transaction: %w", err)
	}

	currentPos, err := s.fileManager.GetCurrentPosition()
	if err != nil {
		return fmt.Errorf("error getting current position: %w", err)
	}

	if err := s.fileManager.SaveProgress(currentPos + 1); err != nil {
		return fmt.Errorf("error saving progress: %w", err)
	}

	return nil
}

func (s *CSVStore) parseTransaction(record []string) (Transaction, error) {
	if len(record) != len(s.headers.Names) {
		return Transaction{}, fmt.Errorf("invalid record length: expected 5, got %d", len(record))
	}

	amount, err := strconv.ParseFloat(record[s.headers.AmtPos], 64)

	if err != nil {
		return Transaction{}, fmt.Errorf("invalid amount: %w", err)
	}

	split, err := strconv.ParseFloat(record[s.headers.SplitPos], 64)
    if err != nil {
        fmt.Printf("Invalid split input: error converting %s: %v\n", record[s.headers.SplitPos], err)
        return Transaction{}, err
    }

	return Transaction{
		Date:        record[s.headers.DatePos],
		Description: record[s.headers.DescPos],
		Amount:      amount,
		Category:    Category(record[s.headers.CatPos]), //TODO extension of band aid
		Split:       split,
	}, nil
}
