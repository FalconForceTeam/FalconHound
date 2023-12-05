package input_processor

import (
	"bytes"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type LogScaleConfig struct {
}

type LogScaleProcessor struct {
	*InputProcessor
	Config LogScaleConfig
}

func (m *LogScaleProcessor) ExecuteQuery() (internal.QueryResults, error) {
	results, err := queryLogscale(m.Query, m.Credentials, m.Debug)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected HTTP status code: 400") {
			return internal.QueryResults{}, fmt.Errorf("failed to run query %q. Most likely there is a syntax error in the query", m.Query)
		} else {
			return internal.QueryResults{}, fmt.Errorf("failed to run query %q: %w", m.Query, err)
		}
	}

	var LogScaleResults internal.QueryResults

	err = json.Unmarshal([]byte(results), &LogScaleResults)
	if err != nil {
		return internal.QueryResults{}, fmt.Errorf("failed to unmarshal JSON from LogScale: %v", err)
	}

	return LogScaleResults, nil
}

func queryLogscale(query string, credentials internal.Credentials, debug bool) (string, error) {
	payload := map[string]interface{}{
		"queryString": query,
		"start":       "15m",
		"end":         "now",
		"isLive":      false,
	}
	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/repositories/%s/query", credentials.LogScaleUrl, credentials.LogScaleRepository), bytes.NewBuffer(jsonPayload))
	req.Header.Set("Authorization", "Bearer "+credentials.LogScaleToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Println(string(bodyBytes))

	return string(bodyBytes), nil
}
