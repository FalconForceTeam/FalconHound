package output_processor

import (
	"bufio"
	"falconhound/internal"
	"fmt"
	"os"
	"strings"
	"time"
)

type MDOutputConfig struct {
	Path             string
	QueryName        string
	QueryEventID     string
	QueryDescription string
	QueryDate        string
}

type MDOutputProcessor struct {
	*OutputProcessor
	Config MDOutputConfig
}

// MD does not require batching, will write all output in one go
func (m *MDOutputProcessor) BatchSize() int {
	return 0
}

func (m *MDOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	err := WriteMD(QueryResults, m.Config)
	return err
}

// WriteMD writes the results to a MD file
func WriteMD(results internal.QueryResults, config MDOutputConfig) error {
	//replace {{date}} with the current date if it exists
	path := strings.Replace(config.Path, "{{date}}", time.Now().Format("2006-01-02"), 2)
	// create the folder if it doesn't exist
	err := os.MkdirAll(path[:strings.LastIndex(path, "/")], 0755)
	if err != nil {
		return fmt.Errorf("failed creating folder: %w", err)
	}
	// Create a file for writing
	MDFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed creating file: %w", err)
	}
	MDWriter := bufio.NewWriter(MDFile)

	var headers []string
	if len(results) == 0 {
		return nil
	}
	for key := range results[0] {
		headers = append(headers, key)
	}

	table := fmt.Sprintf("# Results for query: %s\n\n", config.QueryEventID)
	table += fmt.Sprintf("## %s\n\n", config.QueryName)
	table += fmt.Sprintf("Description: %s\n", config.QueryDescription)
	table += fmt.Sprintf("Date: %s\n\n", config.QueryDate)
	table += "| " + strings.Join(headers, " | ") + " |\n"
	table += "| " + strings.Repeat("--- | ", len(headers)) + "\n"
	_, err = MDWriter.WriteString(table)
	if err != nil {
		return fmt.Errorf("failed writing to file: %w", err)
	}

	// Generate the table rows
	for _, row := range results {
		var values []string
		for _, header := range headers {
			value := fmt.Sprintf("%v", row[header])
			values = append(values, value)
		}
		tablerow := "| " + strings.Join(values, " | ") + " |\n"
		_, err = MDWriter.WriteString(tablerow)
		if err != nil {
			return fmt.Errorf("failed writing row to file: %w", err)
		}
	}

	MDWriter.Flush()
	MDFile.Close()

	return nil
}
