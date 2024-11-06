package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

type FileManager struct {
	InputPath     string
	OutputPath    string
	ProgressPath  string
	LastPosition  int
	headerWritten bool
}

func NewFileManager(inputPath string, outputPath string) *FileManager {
	progressPath := outputPath + ".progress"
	return &FileManager{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		ProgressPath: progressPath,
	}
}

func (fm *FileManager) Initialize() error {
	if _, err := os.Stat(fm.OutputPath); os.IsNotExist(err) {
		if err := fm.createOutputFile(); err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
	}

	if err := fm.loadProgress(); err != nil {
		return fmt.Errorf("failed to load progress: %w", err)
	}

	return nil
}

func (fm *FileManager) createOutputFile() error {
	file, err := os.Create(fm.OutputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"Date", "Description", "Amount", "Split", "Category", "SplitID"}); err != nil {
		return err
	}
	writer.Flush()
	fm.headerWritten = true
	return writer.Error()
}

func (fm *FileManager) loadProgress() error {
	if _, err := os.Stat(fm.ProgressPath); !os.IsNotExist(err) {
		content, err := os.ReadFile(fm.ProgressPath)
		if err != nil {
			return err
		}
		fmt.Sscanf(string(content), "%d", &fm.LastPosition)
	}
	return nil
}

func (fm *FileManager) saveProgress(position int) error {
	return os.WriteFile(fm.ProgressPath, []byte(fmt.Sprintf("%d", position)), 0644)
}
