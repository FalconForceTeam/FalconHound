package input_processor

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type BHConfig struct {
}

type BHProcessor struct {
	*InputProcessor
	Config BHConfig
}

func (m *BHProcessor) ExecuteQuery() (internal.QueryResults, error) {
	results, err := BHRequest(m.Query, m.Credentials)
	if err != nil {
		return internal.QueryResults{}, err
	}
	return results, nil
}

func BHRequest(query string, creds internal.Credentials) (internal.QueryResults, error) {

	// Convert query from a multiline string from the yaml to a single line string so the API can parse it
	query = strings.ReplaceAll(query, "\n", " ")

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
		return nil, err
	}

	req.Header.Set("User-Agent", internal.Version)
	req.Header.Set("Authorization", fmt.Sprintf("bhesignature %s", creds.BHTokenID))
	req.Header.Set("RequestDate", datetimeFormatted)
	req.Header.Set("Signature", base64.StdEncoding.EncodeToString(digester.Sum(nil)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	fmt.Println("Response:", string(respbody))
	// TODO parse response body into QueryResults
	return nil, nil
}

type Node struct {
	Label    string `json:"label"`
	ObjectId string `json:"objectId"`
}

type Response struct {
	Data struct {
		Nodes map[string]Node `json:"nodes"`
	} `json:"data"`
}

type Output struct {
	Name     string `json:"name"`
	ObjectId string `json:"objectId"`
}

func parseResponse(respbody []byte) (string, error) {
	var response Response
	err := json.Unmarshal(respbody, &response)
	if err != nil {
		return "", err
	}

	var output []Output
	for _, node := range response.Data.Nodes {
		output = append(output, Output{Name: node.Label, ObjectId: node.ObjectId})
	}

	outputJSON, err := json.Marshal(output)
	if err != nil {
		return "", err
	}

	return string(outputJSON), nil
}

func GetBHCacheData(creds internal.Credentials, object string) (string, error) {
	// Convert query from a multiline string from the yaml to a single line string so the API can parse it
	// query = strings.ReplaceAll(query, "\n", " ")
	query := fmt.Sprintf("MATCH (c:%s) RETURN c", object) // not possible to return less yet (c.name AS name,c.objectid as objectid)

	method := "POST"
	uri := "/api/v2/graphs/cypher"
	queryBody := fmt.Sprintf("{\"query\":\"%s\"}", query)
	body := []byte(queryBody)

	digester := hmac.New(sha256.New, []byte(creds.BHTokenKey))
	digester.Write([]byte(fmt.Sprintf("%s%s", method, uri)))
	digester = hmac.New(sha256.New, digester.Sum(nil))
	datetimeFormatted := time.Now().Format("2006-01-02T15:04:05.999999-07:00")
	digester.Write([]byte(datetimeFormatted[:13]))

	digester = hmac.New(sha256.New, digester.Sum(nil))

	if body != nil {
		digester.Write(body)
	}

	bhendpoint := fmt.Sprintf("%s%s", creds.BHUrl, uri)

	req, err := http.NewRequest(method, bhendpoint, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	req.Header.Set("User-Agent", internal.Version)
	req.Header.Set("Authorization", fmt.Sprintf("bhesignature %s", creds.BHTokenID))
	req.Header.Set("RequestDate", datetimeFormatted)
	req.Header.Set("Signature", base64.StdEncoding.EncodeToString(digester.Sum(nil)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	results, err := parseResponse(respbody)
	if err != nil {
		panic(err)
	}

	return results, nil
}
