package output_processor

import (
	"encoding/csv"
	"falconhound/internal"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type CSVOutputConfig struct {
	Path string
}

type CSVOutputProcessor struct {
	*OutputProcessor
	Config CSVOutputConfig
}

// CSV does not require batching, will write all output in one go
func (m *CSVOutputProcessor) BatchSize() int {
	return 0
}

func (m *CSVOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	err := WriteCSV(QueryResults, m.Config.Path)
	return err
}

// WriteCSV writes the results to a CSV file
func WriteCSV(results internal.QueryResults, path string) error {
	//replace {{date}} with the current date if it exists
	path = strings.Replace(path, "{{date}}", time.Now().Format("2006-01-02"), 1)
	// create the folder if it doesn't exist
	err := os.MkdirAll(path[:strings.LastIndex(path, "/")], 0755)
	if err != nil {
		return fmt.Errorf("failed creating folder: %w", err)
	}
	// Create a file for writing
	csvFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed creating file: %w", err)
	}
	// Initialize the writer
	csvWriter := csv.NewWriter(csvFile)
	// Get the keys from the first record
	var keys []string
	if len(results) > 0 {
		for k := range results[0] {
			keys = append(keys, k)
		}
	}
	// Sort the keys for consistency between runs
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	// Write the header
	if err := csvWriter.Write(keys); err != nil {
		return fmt.Errorf("failed writing header: %w", err)
	}
	for _, record := range results {
		var row []string
		for _, k := range keys {
			v, ok := record[k]
			if !ok {
				v = nil
			}
			row = append(row, fmt.Sprintf("%v", v))
		}
		err := csvWriter.Write(row)
		if err != nil {
			return fmt.Errorf("failed writing row: %w", err)
		}
	}
	// Flush memory to disk
	csvWriter.Flush()
	csvFile.Close()
	return nil
}
