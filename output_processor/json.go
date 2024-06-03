package output_processor

import (
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type JSONOutputConfig struct {
	Path string
}

type JSONOutputProcessor struct {
	*OutputProcessor
	Config JSONOutputConfig
}

// JSON does not require batching, will write all output in one go
func (m *JSONOutputProcessor) BatchSize() int {
	return 0
}

func (m *JSONOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	err := WriteJSON(QueryResults, m.Config.Path)
	return err
}

func WriteJSON(results internal.QueryResults, path string) error {
	path = strings.Replace(path, "{{date}}", time.Now().Format("2006-01-02"), 2)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed creating directories: %w", err)
	}

	jsonFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed creating file: %w", err)
	}
	defer jsonFile.Close()

	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed marshalling data: %w", err)
	}

	_, err = jsonFile.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed writing data: %w", err)
	}

	return nil
}
