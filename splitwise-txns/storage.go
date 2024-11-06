package main

import (
	"encoding/csv"
	"fmt"
	"io"
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
	ReadTransactions(isRawData bool) ([]Transaction, error)
	SaveTransaction(transaction *Transaction) error
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

func (s *CSVStore) ReadTransactions(isRawData bool) ([]Transaction, error) {
	file, err := os.Open(s.fileManager.InputPath)

	if err != nil {
		return nil, fmt.Errorf("error opening new ile %w", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)

	if _, err := reader.Read(); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("file is empty")
		}
		return nil, fmt.Errorf("error reading header: %w", err)
	}

	for i := 0; i < s.fileManager.LastPosition; i++ {
        _, err := reader.Read()
        if err != nil {
            if err == io.EOF {
                return nil, fmt.Errorf("no new transactions after position %d", s.fileManager.LastPosition)
            }
            return nil, fmt.Errorf("error skipping to position %d: %w", s.fileManager.LastPosition, err)
        }
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

		transaction, err := s.parseTransaction(record, isRawData)

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
	fmt.Println(position)

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

func (s *CSVStore) parseTransaction(record []string, isRawData bool) (Transaction, error) {
	amount, err := strconv.ParseFloat(record[s.headers.AmtPos], 64)

	if err != nil {
		return Transaction{}, fmt.Errorf("invalid amount: %w", err)
	}

	split := 0.0

	if !isRawData { 
		split, err = strconv.ParseFloat(record[s.headers.SplitPos], 64)
		if err != nil {
			fmt.Printf("Invalid split input: error converting %s: %v\n", record[s.headers.SplitPos], err)
			return Transaction{}, err
		}
	} 

	return Transaction{
		Date:        record[s.headers.DatePos],
		Description: record[s.headers.DescPos],
		Amount:      amount,
		Category:    Category(record[s.headers.CatPos]), //TODO extension of band aid
		Split:       split,
	}, nil
}

func (s *CSVStore) SaveTransaction(transaction *Transaction) error {
    dir := filepath.Dir(s.fileManager.OutputPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("error creating directory %s: %w", dir, err)
    }

    if _, err := os.Stat(s.fileManager.OutputPath); os.IsNotExist(err) {
        if err := s.createOutputFileWithHeaders(); err != nil {
            return fmt.Errorf("error creating output file: %w", err)
        }
    }

    existingTransactions, err := s.readAllTransactions()
    if err != nil {
        return fmt.Errorf("error reading existing transactions: %w", err)
    }

    isDuplicate := false
    for _, existing := range existingTransactions {
        if isMatchingTransaction(&existing, transaction) {
            isDuplicate = true
            break
        }
    }

    if !isDuplicate {
        existingTransactions = append(existingTransactions, *transaction)
    }

    if err := s.writeAllTransactions(existingTransactions); err != nil {
        return fmt.Errorf("error writing transactions: %w", err)
    }

    if !isDuplicate {
        currentPos, err := s.fileManager.GetCurrentPosition()
        if err != nil {
            return fmt.Errorf("error getting current position: %w", err)
        }
        if err := s.fileManager.SaveProgress(currentPos + 1); err != nil {
            return fmt.Errorf("error saving progress: %w", err)
        }
    }

    return nil
}

func isMatchingTransaction(a *Transaction, b *Transaction) bool {
    return a.Date == b.Date &&
        a.Description == b.Description &&
        a.Amount == b.Amount
}

func (s *CSVStore) readAllTransactions() ([]Transaction, error) {
    file, err := os.Open(s.fileManager.OutputPath)
    if err != nil {
        return nil, fmt.Errorf("error opening file: %w", err)
    }
    defer file.Close()

    reader := csv.NewReader(file)
    
    if _, err := reader.Read(); err != nil {
        if err == io.EOF {
            return []Transaction{}, nil
        }
        return nil, fmt.Errorf("error reading header: %w", err)
    }

    var transactions []Transaction
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("error reading record: %w", err)
        }

        transaction, err := s.parseTransaction(record, false)
        if err != nil {
            return nil, fmt.Errorf("error parsing record: %w", err)
        }

        transactions = append(transactions, transaction)
    }

    return transactions, nil
}

func (s *CSVStore) writeAllTransactions(transactions []Transaction) error {
    file, err := os.Create(s.fileManager.OutputPath)
    if err != nil {
        return fmt.Errorf("error creating file: %w", err)
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    if err := writer.Write([]string{"Date", "Description", "Amount", "Category", "Split"}); err != nil {
        return fmt.Errorf("error writing header: %w", err)
    }

    for _, txn := range transactions {
        record := []string{
            txn.Date,
            txn.Description,
            fmt.Sprintf("%.2f", txn.Amount),
            string(txn.Category),
            fmt.Sprintf("%.2f", txn.Split),
        }
        if err := writer.Write(record); err != nil {
            return fmt.Errorf("error writing record: %w", err)
        }
    }

    return nil
}
