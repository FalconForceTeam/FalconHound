package input_processor

import (
	"crypto/tls"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type JobResponse struct {
	Sid string `json:"sid"`
}

type JobStatusResponse struct {
	Entry []struct {
		Content struct {
			IsDone bool `json:"isDone"`
		} `json:"content"`
	} `json:"entry"`
}

type SplunkResults struct {
	Schema []struct {
		Name string `json:"Name"`
		Type string `json:"Type"`
	} `json:"Schema"`
	Results internal.QueryResults `json:"results"`
}

type SplunkConfig struct {
}

type SplunkProcessor struct {
	*InputProcessor
	Config SplunkConfig
}

func (m *SplunkProcessor) ExecuteQuery() (internal.QueryResults, error) {
	results, err := QuerySplunk(m.Query, m.Credentials, m.Debug)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected HTTP status code: 400") {
			return internal.QueryResults{}, fmt.Errorf("failed to run query %q. Most likely there is a syntax error in the query", m.Query)
		} else {
			return internal.QueryResults{}, fmt.Errorf("failed to run query %q: %w", m.Query, err)
		}
	}

	// Get rows
	var SplunkResults SplunkResults

	err = json.Unmarshal([]byte(results), &SplunkResults)
	if err != nil {
		return internal.QueryResults{}, fmt.Errorf("failed to unmarshal JSON from Splunk: %v", err)
	}

	queryResults := SplunkResults.Results

	return queryResults, nil
}

func QuerySplunk(query string, credentials internal.Credentials, debug bool) (string, error) {
	baseURL := credentials.SplunkUrl + ":" + credentials.SplunkApiPort
	authHeader := "Bearer " + credentials.SplunkApiToken

	// Create a custom HTTP client and ignore SSL errors
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	searchQuery := "search " + query
	// replace %s in query with the credentials.SplunkIndex
	searchQuery = strings.Replace(searchQuery, "%s", credentials.SplunkIndex, 1)

	// Start a search job and request json response
	data := url.Values{}
	data.Set("search", searchQuery)
	data.Set("output_mode", "json")

	req, err := http.NewRequest("POST", baseURL+"/services/search/jobs", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var jobResponse JobResponse
	err = json.Unmarshal(bodyBytes, &jobResponse)
	if err != nil {
		log.Fatalln(err)
	}

	sid := jobResponse.Sid

	polldata := url.Values{}
	polldata.Set("output_mode", "json")

	// Poll for search job completion
	for {
		req, err := http.NewRequest("GET", baseURL+"/services/search/jobs/"+sid, strings.NewReader(polldata.Encode()))
		if err != nil {
			log.Fatalln(err)
		}
		req.Header.Set("Authorization", authHeader)

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		var jobStatusResponse JobStatusResponse
		err = json.Unmarshal(bodyBytes, &jobStatusResponse)
		if err != nil {
			log.Fatalln(err)
		}

		if jobStatusResponse.Entry[0].Content.IsDone {
			if debug {
				log.Printf("[i] Job %s completed, getting data\n", sid)
			}
			break
		}

		if debug {
			log.Printf("[»] Job %s still running, waiting 1 second\n", sid)
		}
		time.Sleep(1 * time.Second)
	}

	// Get the search results
	resultdata := url.Values{}
	resultdata.Set("output_mode", "json")
	req, err = http.NewRequest("GET", baseURL+"/services/search/jobs/"+sid+"/results", strings.NewReader(resultdata.Encode()))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Authorization", authHeader)

	resp, err = client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return string(bodyBytes), nil

}
