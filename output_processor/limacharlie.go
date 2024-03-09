package output_processor

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"net/http"
)

type LimaCharlieOutputConfig struct {
	BHQuery          string
	QueryName        string
	QueryDescription string
	QueryEventID     string
}

type LimaCharlieOutputProcessor struct {
	*OutputProcessor
	Config LimaCharlieOutputConfig
}

func (m *LimaCharlieOutputProcessor) BatchSize() int {
	return 0
}

func (m *LimaCharlieOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	jsonData, err := json.Marshal(QueryResults)
	if err != nil {
		return err
	}
	LimaCharlieData := map[string]interface{}{
		"Name":        m.Config.QueryName,
		"Description": m.Config.QueryDescription,
		"EventID":     m.Config.QueryEventID,
		"BHQuery":     m.Config.BHQuery,
		"EventData":   string(jsonData),
	}
	// Wrap LimaCharlieData in JSON
	LimaCharlieJSONData, err := json.Marshal(LimaCharlieData)
	if err != nil {
	}

	url := (m.Credentials.LimaCharlieAPIUrl)
	fmt.Println(url)
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(LimaCharlieJSONData))
	if err != nil {
		return err
	}

	// Set headers
	req.SetBasicAuth(m.Credentials.LimaCharlieOrgId, m.Credentials.LimaCharlieIngestKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("lc-source", "FalconHound")
	req.Header.Set("lc-hint", "json")

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
