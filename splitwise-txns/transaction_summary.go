package main

import (
	"fmt"
	//	"path/filepath"
	"sort"
	"strings"
)

type CategorySummary struct {
	TotalAmount      float64
	TransactionCount int
}

type SummaryReport struct {
	Categories  map[Category]CategorySummary
	TotalAmount float64
}

type SummarizeProcessor struct {
	projectName string
	store       TransactionStore
}

func NewSummarizeProcessor(projectName string) *SummarizeProcessor {
	fm := NewFileManager(projectName, "")
	csvStore := NewCSVStore(fm)

	return &SummarizeProcessor{
		projectName: projectName,
		store:       csvStore,
	}
}

func (ts *SummarizeProcessor) Run(args []string) error {
	rawData := false
	transactions, err := ts.store.ReadTransactions(rawData)

	if err != nil {
		return fmt.Errorf("error reading transactions %w", err)
	}

	report := Summarize(transactions)
	formatter := NewSummaryFormatter()

	formattedReport := formatter.FormatReport(report)
	fmt.Print(formattedReport)

	return nil
}

// helper TODO is this right?
func Summarize(transactions []Transaction) *SummaryReport {
	summaries := make(map[Category]CategorySummary)
	var totalAmount float64

	for _, txn := range transactions {
		summary := summaries[txn.Category]
		summary.TotalAmount += (txn.Split / 100) * txn.Amount
		summary.TransactionCount++
		summaries[txn.Category] = summary
		totalAmount += txn.Amount
	}

	return &SummaryReport{
		Categories:  summaries,
		TotalAmount: totalAmount,
	}
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
