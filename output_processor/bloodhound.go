package output_processor

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"falconhound/internal"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type BloodHoundOutputConfig struct {
	Query      string
	Parameters map[string]string
}

type BloodHoundOutputProcessor struct {
	*OutputProcessor
	Config BloodHoundOutputConfig
}

func (m *BloodHoundOutputProcessor) BatchSize() int {
	return 1
}

func (m *BloodHoundOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
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

	return WriteBloodHound(m.Config.Query, params, m.Credentials)
}

// TODO also embed the driver and session in the struct
//var session BloodHound.Session

func WriteBloodHound(query string, params map[string]interface{}, creds internal.Credentials) error {
	if creds.BHTokenKey == "" {
		return fmt.Errorf("BHTokenKey is empty, skipping..")
	}

	// replace parameters in query
	//for key, value := range params {
	//	query = strings.ReplaceAll(query, fmt.Sprintf("$%s", key), fmt.Sprintf("'%v'", value))
	//}
	for key, value := range params {
		upperValue := strings.ToUpper(fmt.Sprintf("%v", value))
		query = strings.ReplaceAll(query, fmt.Sprintf("$%s", key), fmt.Sprintf("'%s'", upperValue))
	}

	// Convert query from a multiline string from the yaml to a single line string so the API can parse it
	query = strings.ReplaceAll(query, "\n", " ")
	log.Printf("Query: %s\n", query)

	method := "POST"
	uri := "/api/v2/graphs/cypher"
	queryBody := fmt.Sprintf("{\"query\":\"%s\"}", query)
	body := []byte(queryBody)

	// The first HMAC digest is the token key
	digester := hmac.New(sha256.New, []byte(creds.BHTokenKey))

	// OperationKey is the first HMAC digestresource
	digester.Write([]byte(fmt.Sprintf("%s%s", method, uri)))

	// Update the digester for further chaining
	digester = hmac.New(sha256.New, digester.Sum(nil))
	datetimeFormatted := time.Now().Format("2006-01-02T15:04:05.999999-07:00")
	digester.Write([]byte(datetimeFormatted[:13]))

	// Update the digester for further chaining
	digester = hmac.New(sha256.New, digester.Sum(nil))

	// Body signing is the last HMAC digest link in the signature chain. This encodes the request body as part of
	// the signature to prevent replay attacks that seek to modify the payload of a signed request. In the case
	// where there is no body content the HMAC digest is computed anyway, simply with no values written to the
	// digester.
	if body != nil {
		digester.Write(body)
	}

	bhendpoint := fmt.Sprintf("%s%s", creds.BHUrl, uri)

	// Perform the request with the signed and expected headers
	req, err := http.NewRequest(method, bhendpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", internal.Version)
	req.Header.Set("Authorization", fmt.Sprintf("bhesignature %s", creds.BHTokenID))
	req.Header.Set("RequestDate", datetimeFormatted)
	req.Header.Set("Signature", base64.StdEncoding.EncodeToString(digester.Sum(nil)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	fmt.Println("Response:", string(respbody))
	// TODO parse response body into QueryResults
	return nil
}
