package output_processor

import (
	"bytes"
	"context"
	"encoding/json"
	"falconhound/internal"
	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"log"
	"time"
)

type ADXOutputConfig struct {
	BHQuery          string
	QueryName        string
	QueryDescription string
	QueryEventID     string
	Table            string
	BatchSize        int
}

type ADXOutputProcessor struct {
	*OutputProcessor
	Config ADXOutputConfig
}

func (m *ADXOutputProcessor) BatchSize() int {
	if m.Config.BatchSize > 0 {
		return m.Config.BatchSize
	}
	return 1
}

func (m *ADXOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	jsonData, err := json.Marshal(QueryResults)
	if err != nil {
		log.Fatalf("failed to marshal data: %s", err)
		return err
	}

	runTime := time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z07:00")
	// Create a data object ADXdata with data from EventData or EnrichmentData, whichever is not nil
	ADXData := map[string]interface{}{
		"Name":        m.Config.QueryName,
		"Description": m.Config.QueryDescription,
		"EventID":     m.Config.QueryEventID,
		"BHQuery":     m.Config.BHQuery,
		"Timestamp":   runTime,
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
