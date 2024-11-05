package main

import (
    "encoding/csv"
    "fmt"
    "io"
    "os"
    "sort"
    "strconv"
    "strings"
)

type CategorySummary struct {
    TotalAmount      float64
    TransactionCount int
}

type SummaryReport struct {
    Categories   map[Category]CategorySummary
    TotalAmount float64
}

type TransactionSummarizer struct{}

func NewTransactionSummarizer() *TransactionSummarizer {
    return &TransactionSummarizer{}
}

func (ts *TransactionSummarizer) Summarize(transactions []Transaction) *SummaryReport {
    summaries := make(map[Category]CategorySummary)
    var totalAmount float64

    for _, txn := range transactions {
        summary := summaries[txn.Category]
        summary.TotalAmount += txn.Amount
        summary.TransactionCount++
        summaries[txn.Category] = summary
        totalAmount += txn.Amount
    }

    return &SummaryReport{
        Categories:   summaries,
        TotalAmount: totalAmount,
    }
}

type TransactionReader struct {
    filePath string
}

func NewTransactionReader(filePath string) *TransactionReader {
    return &TransactionReader{
        filePath: filePath,
    }
}

func (tr *TransactionReader) ReadTransactions() ([]Transaction, error) {
    file, err := os.Open(tr.filePath)
    if err != nil {
        return nil, fmt.Errorf("error opening file: %w", err)
    }
    defer file.Close()

    reader := csv.NewReader(file)
    
    if _, err := reader.Read(); err != nil {
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

        amount, err := tr.parseAmount(record[2])
        if err != nil {
            return nil, fmt.Errorf("error parsing amount for transaction: %w", err)
        }

		split, err := tr.parseSplit(record[3])

        if err != nil {
            return nil, fmt.Errorf("error parsing splitfor transaction: %w", err)
        }


        transactions = append(transactions, Transaction{
            Date:        record[0],
            Description: record[1],
            Amount:      amount,
			Split:       split,
            Category:    Category(record[4]),
        })
    }

    return transactions, nil
}

func (tr *TransactionReader) parseAmount(amountStr string) (float64, error) {
    cleaned := strings.TrimSpace(strings.ReplaceAll(amountStr, "$", ""))
    cleaned = strings.ReplaceAll(cleaned, ",", "")
    return strconv.ParseFloat(cleaned, 64)
}

func (tr *TransactionReader) parseSplit(amountStr string) (int, error) {
	intVal, err := strconv.Atoi(str)
    if err != nil {
        fmt.Println("Error converting string to int:", err)
        return nil, err
    }
    return intVal, nil
}

type SummaryFormatter struct{}

func NewSummaryFormatter() *SummaryFormatter {
    return &SummaryFormatter{}
}

func (sf *SummaryFormatter) FormatReport(report *SummaryReport) string {
    var sb strings.Builder

    sb.WriteString("\nTransaction Summary\n")
    sb.WriteString("==================\n\n")
    sb.WriteString(fmt.Sprintf("Total Amount: $%.2f\n\n", report.TotalAmount))
    sb.WriteString("Breakdown by Category:\n")
    sb.WriteString("---------------------\n")

    categories := sf.getSortedCategories(report.Categories)

    for _, category := range categories {
        summary := report.Categories[category]
        percentage := (summary.TotalAmount / report.TotalAmount) * 100
        
        sb.WriteString(fmt.Sprintf("%-12s: $%10.2f (%6.2f%%) [%d transactions]\n",
            category,
            summary.TotalAmount,
            percentage,
            summary.TransactionCount))
    }

    return sb.String()
}

func (sf *SummaryFormatter) getSortedCategories(summaries map[Category]CategorySummary) []Category {
    var categories []Category
    for category := range summaries {
        categories = append(categories, category)
    }
    sort.Slice(categories, func(i, j int) bool {
        return string(categories[i]) < string(categories[j])
    })
    return categories
}

func handleSummarize(filePath string) error {
    reader := NewTransactionReader(filePath)
    summarizer := NewTransactionSummarizer()
    formatter := NewSummaryFormatter()

    transactions, err := reader.ReadTransactions()
    if err != nil {
        return fmt.Errorf("error reading transactions: %w", err)
    }

    report := summarizer.Summarize(transactions)

    formattedReport := formatter.FormatReport(report)
    fmt.Print(formattedReport)

    return nil
}
