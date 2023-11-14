package output_processor

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"net/http"
)

type SplunkOutputConfig struct {
}

type SplunkOutputProcessor struct {
	*OutputProcessor
	Config SplunkOutputConfig
}

func (m *OutputProcessor) BatchSize() int {
	return 0
}

func (m *SplunkOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	return PostToSplunk(QueryResults, m.Credentials)
}

func PostToSplunk(queryResults internal.QueryResults, creds internal.Credentials) error {
	// Wrap data in JSON
	payload := map[string]internal.QueryResults{
		"event": queryResults,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s:%s/services/collector/event", creds.SplunkUrl, creds.SplunkHecPort)
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Splunk %s", creds.SplunkHecToken))
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client with custom transport to disable SSL verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Send HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
