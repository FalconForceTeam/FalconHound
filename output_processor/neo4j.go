package output_processor

import (
	"fmt"

	"falconhound/internal"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Neo4jOutputConfig struct {
	Query      string
	Parameters map[string]string
}

type Neo4jOutputProcessor struct {
	*OutputProcessor
	Config Neo4jOutputConfig
}

func (m *Neo4jOutputProcessor) BatchSize() int {
	return 1
}

func (m *Neo4jOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	if len(QueryResults) == 0 {
		return nil
	}
	var queryResult internal.QueryResult = QueryResults[0]
	var params = make(map[string]interface{})
	for key, value := range m.Config.Parameters {
		rowValue, ok := queryResult[value]
		if !ok {
			return fmt.Errorf("parameter %s not found in query results", value)
		}
		// Insert into map
		params[key] = rowValue
	}
	if m.Debug {
		fmt.Printf("Query: %#v, parameters: %#v\n", m.Config.Query, params)
	}

	return WriteNeo4j(m.Config.Query, params, m.Credentials)
}

// TODO also embed the driver and session in the struct
var session neo4j.Session

func WriteNeo4j(cypher string, params map[string]interface{}, creds internal.Credentials) error {
	// Create a new Neo4j driver
	driver, err := neo4j.NewDriver(creds.Neo4jUri, neo4j.BasicAuth(creds.Neo4jUsername, creds.Neo4jPassword, ""))
	if err != nil {
		return fmt.Errorf("error connecting to Neo4j: %w", err)
	}

	// Create a new Neo4j session if there is no current one
	if session == nil {
		session = driver.NewSession(neo4j.SessionConfig{})
	}

	// Execute the Cypher query using the session
	_, err = session.Run(cypher, params)
	if err != nil {
		return fmt.Errorf("error executing cypher query: %w", err)
	}
	return nil
}

func Finalize() error {
	if session == nil {
		return nil
	}
	err := session.Close()
	session = nil
	return err
}
