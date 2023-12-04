package output_processor

import (
	"bytes"
	"context"
	"encoding/json"
	"falconhound/internal"
	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"log"
)

type ADXOutputConfig struct {
	BHQuery          string
	QueryName        string
	QueryDescription string
	QueryEventID     string
	Table            string
}

type ADXOutputProcessor struct {
	*OutputProcessor
	Config ADXOutputConfig
}

func (m *ADXOutputProcessor) BatchSize() int {
	return 1
}

func (m *ADXOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	jsonData, err := json.Marshal(QueryResults)
	if err != nil {
		log.Fatalf("failed to marshal data: %s", err)
		return err
	}

	// Create a data object ADXdata with data from EventData or EnrichmentData, whichever is not nil
	ADXData := map[string]interface{}{
		"Name":        m.Config.QueryName,
		"Description": m.Config.QueryDescription,
		"EventID":     m.Config.QueryEventID,
		"BHQuery":     m.Config.BHQuery,
		"EventData":   string(jsonData),
	}

	kustoConnectionStringBuilder := kusto.NewConnectionStringBuilder(m.Credentials.AdxClusterURL)
	kustoConnectionString := kustoConnectionStringBuilder.WithAadAppKey(m.Credentials.AdxAppID, m.Credentials.AdxAppSecret, m.Credentials.AdxTenantID)

	client, err := kusto.New(kustoConnectionString)
	if err != nil {
		log.Fatalf("failed to create Kusto client: %s", err)
		return err
	}

	// Create an ingestion client
	ingestor, err := ingest.New(client, m.Credentials.AdxDatabase, m.Config.Table)
	if err != nil {
		log.Fatalf("failed to create ingestor: %s", err)
		return err
	}

	if m.Debug {
		log.Printf("clusterURL: %s", m.Credentials.AdxClusterURL)
		log.Println("Ingesting the following data: ", ADXData)
	}

	// Ingest the data
	data, err := json.Marshal(ADXData)
	ctx := context.Background()
	reader := bytes.NewReader(data)
	_, err = ingestor.FromReader(ctx, reader, ingest.FileFormat(ingest.MultiJSON))
	if err != nil {
		log.Fatalf("failed to create file: %s", err)
		return err
	}

	return client.Close()
}
