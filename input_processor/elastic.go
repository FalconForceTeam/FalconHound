package input_processor

import (
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"io"
	"log"
	"strings"
)

type ElasticConfig struct {
}

type ElasticProcessor struct {
	*InputProcessor
	Config ElasticConfig
}

type ElasticsearchResponse struct {
	Hits struct {
		Hits []struct {
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (m *ElasticProcessor) ExecuteQuery() (internal.QueryResults, error) {
	if m.Credentials.ElasticApiKey == "" {
		return internal.QueryResults{}, fmt.Errorf("ElasticApiKey is empty, skipping..")
	}

	results, err := queryElastic(m.Query, m.Credentials, m.Debug)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected HTTP status code: 400") {
			return internal.QueryResults{}, fmt.Errorf("failed to run query %q. Most likely there is a syntax error in the query", m.Query)
		} else {
			return internal.QueryResults{}, fmt.Errorf("failed to run query %q: %w", m.Query, err)
		}
	}

	var ElasticResults internal.QueryResults

	err = json.Unmarshal([]byte(results), &ElasticResults)
	if err != nil {
		return internal.QueryResults{}, fmt.Errorf("failed to unmarshal JSON from Elastic: %v", err)
	}

	return ElasticResults, nil
}

func queryElastic(query string, credentials internal.Credentials, debug bool) (string, error) {

	cfg := elasticsearch.Config{
		CloudID: credentials.ElasticCloudID,
		APIKey:  credentials.ElasticApiKey,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	cleanquery := strings.ReplaceAll(query, "\n", " ")
	esquery := fmt.Sprintf(`{ "query": { "query_string": { "query": "(%s)" }}}`, cleanquery)
	fmt.Printf("Query: %s\n", esquery)
	res, err := es.Search(
		es.Search.WithIndex(".ds-logs-system.security-default*"), // <-- TODO: now with index
		es.Search.WithBody(strings.NewReader(esquery)),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()
	//log.Println(res)

	bodyBytes, _ := io.ReadAll(res.Body)

	var esResponse ElasticsearchResponse
	if err := json.Unmarshal([]byte(bodyBytes), &esResponse); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		//return
	}

	var allFlattenedData []map[string]interface{}

	for _, hit := range esResponse.Hits.Hits {
		var result map[string]interface{}
		err := json.Unmarshal(hit.Source, &result)
		if err != nil {
			log.Fatalf("Error unmarshalling JSON: %v", err)
		}

		flattenedData := make(map[string]interface{})
		flatten(result, "", flattenedData)

		allFlattenedData = append(allFlattenedData, flattenedData)
	}

	flattenedJSON, err := json.Marshal(allFlattenedData)
	if err != nil {
		log.Fatalf("Error marshalling flattened data: %v", err)
	}

	return string(flattenedJSON), nil
}

func flatten(data interface{}, prefix string, flattenedData map[string]interface{}) {
	switch data := data.(type) {
	case map[string]interface{}:
		for k, v := range data {
			var fullKey string
			if prefix == "" {
				fullKey = k
			} else {
				fullKey = prefix + "." + k
			}
			flatten(v, fullKey, flattenedData)
		}
	case []interface{}:
		for i, v := range data {
			flatten(v, fmt.Sprintf("%s.%d", prefix, i), flattenedData)
		}
	default:
		flattenedData[prefix] = data
	}
}
