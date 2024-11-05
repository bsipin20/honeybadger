
package main

type HeaderConfig struct {
    Names    []string
    DatePos  int
    DescPos  int
    AmtPos   int
    CatPos   int
    SplitPos int
}

var DefaultHeaders = HeaderCOnfig{
	Names: []string{"Date", "Description", "Amount", "Category", "Split"},
	DatePos: 0,
	DescPos: 1,
	AmtPos: 2,
	CatPos: 3,
	SplitPos: 4,
}

type TransactionStore interface {
	ReadTransactions(filename string) ([]Transaction, error)
	WriteTransactions(filename string, transactions []model.Transaction) error
	UpdateTransaction(filename string, id string, transaction model.Transaction)
}

type Transaction struct {
	Date        string
	Description string
	Amount      float64
	Split       int
	Category    Category
	SplitID     string
}

type CSVStore struct {
	headers HeaderConfig

}

//TODO implement option to write files to directory based on project name
func NewCSVStore(headers HeaderConfig) *CSVStore {
	return &CSVStore{
		headers: headers,
	}
}

func (s *CSVStore) ReadTransactions(filename string) ([]Tranmsaction, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening new ile %w", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)


	header, err := reader.Read()

	if err != nil {
		return nil, fmt.Errorf("error reading header: %w", err)
	}

	if err := validateHeader(header); err != nil {
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

		transaction, err := parseTransaction(record)

		if err != nil {
			return nil, fmt.Errorf("error parsing record: %w", err)
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}


func (s *CSVStore) SaveTransactions(filename string, transactions []Transaction) error {
    file, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("error creating file: %w", err)
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    header := []string{"Date", "Description", "Amount", "Category", "Split"}
    if err := writer.Write(header); err != nil {
        return fmt.Errorf("error writing header: %w", err)
    }

    for _, txn := range transactions {
        record := []string{
            txn.Date,
            txn.Description,
            fmt.Sprintf("%.2f", txn.Amount),
            txn.Category,
            fmt.Sprintf("%.2f", txn.Split),
        }
        if err := writer.Write(record); err != nil {
            return fmt.Errorf("error writing record: %w", err)
        }
    }

    return nil
}

func (s *CSVStore) validateHeader(header []string) error {
    if len(header) != len(s.headers.Names) {
        return fmt.Errorf("invalid header length: expected %d, got %d", len(s.headers.Names), len(header))
    }
    for i, field := range s.headers.Names {
        if header[i] != field {
            return fmt.Errorf("invalid header field at position %d: expected %s, got %s", i, field, header[i])
        }
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
		return Transaction{}, fmt.Errorf("invalid split: %w", err)
	}

	return Transaction{
		Date: record[s.headers.DatePos],
		Description: record[s.headers.DescPos],
		Amount: amount,
		Category: record[s.headers.CatPos],
		Split: split,
	}, nil
}

type Processor struct {
	store TransactionStore
}

func NewTransactionProcessor() *Processor {
	return &Processor {
		store: NewCSVStore(),
	}
}


