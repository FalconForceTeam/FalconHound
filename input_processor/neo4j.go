package input_processor

import (
	"encoding/json"
	"falconhound/internal"
	"log"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Neo4jConfig struct {
}

type Neo4jProcessor struct {
	*InputProcessor
	Config Neo4jConfig
}

func (m *Neo4jProcessor) ExecuteQuery() (internal.QueryResults, error) {
	results, err := ReadNeo4j(m.Query, m.Credentials)
	if err != nil {
		return internal.QueryResults{}, err
	}
	return results, nil
}

func ReadNeo4j(query string, creds internal.Credentials) (internal.QueryResults, error) {
	driver, err := neo4j.NewDriver(creds.Neo4jUri, neo4j.BasicAuth(creds.Neo4jUsername, creds.Neo4jPassword, ""))
	if err != nil {
		log.Printf("Error connecting to Neo4j: %v", err)
		panic(err)
	}
	defer driver.Close()

	// Create a new Neo4j session
	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()
	// Create a new node
	neoresult, err := session.Run(query, map[string]interface{}{})
	if err != nil {
		log.Printf("Error creating node: %v", err)
		return nil, err
	}

	// Create a slice to hold the merged JSON objects
	queryResults := make(internal.QueryResults, 0)

	// Iterate over the records returned by the query
	for neoresult.Next() {
		record := neoresult.Record()

		// Iterate over the fields of the record
		for i := 0; i < len(record.Values); i++ {
			// Get the value at the current index
			field := record.Values[i]

			// Convert the field to a JSON string
			fieldJSON, err := json.Marshal(field)
			if err != nil {
				log.Printf("Error marshalling JSON: %v", err)
				return nil, err
			}

			// Decode the JSON string into a map
			var fieldMap map[string]interface{}
			err = json.Unmarshal(fieldJSON, &fieldMap)
			if err != nil {
				log.Printf("Error unmarshalling JSON: %v", err)
				return nil, err
			}

			// Append the map to the mergedJSON slice
			queryResults = append(queryResults, fieldMap)
		}
	}
	return queryResults, nil

}
