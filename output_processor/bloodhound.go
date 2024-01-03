package output_processor

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var SessionTypeMap map[string]int = map[string]int{
	"Interactive":             2,
	"Network":                 3,
	"Batch":                   4,
	"Service":                 5,
	"Proxy":                   6,
	"Unlock":                  7,
	"NetworkCleartext":        8,
	"NewCredentials":          9,
	"RemoteInteractive":       10,
	"CachedInteractive":       11,
	"CachedRemoteInteractive": 12,
	"CachedUnlock":            13,
}

type BloodHoundOutputConfig struct {
	BatchSize  int
	OutputType string
}

type BloodHoundOutputProcessor struct {
	*OutputProcessor
	Config BloodHoundOutputConfig
}

func (m *BloodHoundOutputProcessor) BatchSize() int {
	// If batch size is given in the config, use that
	// otherwise default to 10
	if m.Config.BatchSize > 0 {
		return m.Config.BatchSize
	}
	return 10
}

type BHResponseData struct {
	Id            int                    `json:"id"`
	Status        int                    `json:"status"`
	StatusMessage string                 `json:"status_message"`
	Nodes         map[string]interface{} `json:"nodes"`
}

type BHResponse struct {
	Data BHResponseData `json:"data"`
}

// TODO this stuff should go in some kind of helper class since we will need it many times
func QueryBloodhoundAPI(uri string, method string, body []byte, creds internal.Credentials) (BHResponse, error) {
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
		return BHResponse{}, err
	}

	req.Header.Set("User-Agent", internal.Version)
	req.Header.Set("Authorization", fmt.Sprintf("bhesignature %s", creds.BHTokenID))
	req.Header.Set("RequestDate", datetimeFormatted)
	req.Header.Set("Signature", base64.StdEncoding.EncodeToString(digester.Sum(nil)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return BHResponse{}, err
	}

	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return BHResponse{}, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return BHResponse{}, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	var response BHResponse
	// Empty response is OK for some endpoints
	if len(respbody) == 0 {
		return BHResponse{}, nil
	}
	// fmt.Println("Response body:", string(respbody))
	err = json.Unmarshal(respbody, &response)
	if err != nil {
		return BHResponse{}, err
	}
	return response, nil
}

func (m *BloodHoundOutputProcessor) UploadData(data []byte) error {
	upload_job, err := QueryBloodhoundAPI("/api/v2/file-upload/start", "POST", nil, m.Credentials)
	if err != nil {
		return err
	}
	job_id := upload_job.Data.Id
	_, err = QueryBloodhoundAPI(fmt.Sprintf("/api/v2/file-upload/%d", job_id), "POST", data, m.Credentials)
	if err != nil {
		return err
	}
	_, err = QueryBloodhoundAPI(fmt.Sprintf("/api/v2/file-upload/%d/end", job_id), "POST", nil, m.Credentials)
	if err != nil {
		return err
	}
	return nil
}

//type CypherSearch struct {
//	Query             string `json:"query"`
//	IncludeProperties bool   `json:"include_properties,omitempty"`
//}

//func (m *BloodHoundOutputProcessor) GetComputerSIDS(ComputerNames []string) (map[string]string, error) {
//	// Use Json.Marshal to escape the computer names an put them into an array
//	// ["COMPUTER1", "COMPUTER2"] that we can use in the Cypher query
//	NamesAsJson, err := json.Marshal(ComputerNames)
//	if err != nil {
//		return nil, err
//	}
//	// For now we have to use a separate Cypher query to get the computer names.
//	// Also, this query must return a full computer node, otherwise the query results in an error.
//	query := fmt.Sprintf("MATCH (c:Computer) WHERE c.name IN %s RETURN c", NamesAsJson)
//	cypherSearch := CypherSearch{
//		Query:             query,
//		IncludeProperties: false,
//	}
//	cypherSearchAsJson, err := json.Marshal(cypherSearch)
//	if err != nil {
//		return nil, err
//	}
//
//	query_results, err := QueryBloodhoundAPI("/api/v2/graphs/cypher", "POST", cypherSearchAsJson, m.Credentials)
//	if err != nil {
//		return nil, err
//	}
//	results := make(map[string]string)
//	if query_results.Data.Nodes == nil {
//		return results, nil
//	}
//	for _, node := range query_results.Data.Nodes {
//		// Check if node has both label and objectId fields
//		nodeName, ok := node.(map[string]interface{})["label"].(string)
//		if !ok {
//			continue
//		}
//		nodeObjectId, ok := node.(map[string]interface{})["objectId"].(string)
//		if !ok {
//			continue
//		}
//		results[nodeName] = nodeObjectId
//	}
//
//	return results, err
//}

func (m *BloodHoundOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	// Build a list of unique computer names
	computerNameHash := make(map[string]struct{})
	for _, result := range QueryResults {
		computerName := strings.ToUpper(result["DeviceName"].(string))
		computerNameHash[computerName] = struct{}{}
	}
	computerNames := make([]string, 0, len(computerNameHash))
	for computerName := range computerNameHash {
		computerNames = append(computerNames, computerName)
	}

	// Obtain the computer SIDs for the computer names using a Cypher query
	//computerSIDs, err := m.GetComputerSIDS(computerNames)
	//if err != nil {
	//	return err
	//}
	cachedb, _ := internal.OpenDB("cache.db")
	computerSIDs, _ := internal.GetCachedComputerByNames(cachedb, computerNames)
	//fmt.Println("results from query: ", computerSIDs)
	internal.CloseDB(cachedb)

	switch m.Config.OutputType {
	case "Session":
		return m.UploadData(SessionFactory(QueryResults, computerSIDs))
	case "Computer":
		return m.UploadData(ComputerFactory(QueryResults, computerSIDs))
	}

	//json := SessionFactory(QueryResults, computerSIDs)
	//return m.UploadData(json)
	return nil
}

func SessionFactory(QueryResults internal.QueryResults, computerSIDs map[string]string) []byte {
	sessions := make([]internal.Session, 0)
	for _, result := range QueryResults {
		logonType, ok := SessionTypeMap[result["LogonType"].(string)]
		if !ok {
			logonType = 0
		}
		deviceName := strings.ToUpper(result["DeviceName"].(string))
		computerSID, ok := computerSIDs[deviceName]
		if !ok {
			log.Printf("Warning - could not find computer SID for device %s\n", deviceName)
			continue
		}
		// additional properties sadly not possible https://github.com/SpecterOps/BloodHound/blob/f35ce41e2a41ab39d4614afd7cf1f3f584ae6328/cmd/api/src/daemons/datapipe/ingest.go#L338
		// schema > https://github.com/SpecterOps/BloodHound/blob/main/packages/go/graphschema/ad/ad.go
		session := internal.Session{
			ComputerSID: computerSID,
			UserSID:     result["AccountSid"].(string),
			LogonType:   logonType,
		}
		sessions = append(sessions, session)
	}
	metadata := internal.Metadata{
		Type:    internal.DataTypeSession,
		Methods: internal.CollectionMethodSession,
		Version: 6,
	}
	json, err := FormatBloodHounds(metadata, sessions)
	if err != nil {
		return nil
	}
	fmt.Printf("Uploading %d sessions\n", len(sessions))
	//fmt.Println(string(json))
	return json
}

func ComputerFactory(QueryResults internal.QueryResults, computerSIDs map[string]string) []byte {
	computers := make([]internal.Computer, 0)
	for _, result := range QueryResults {

		deviceName := strings.ToUpper(result["DeviceName"].(string))
		computerSID, ok := computerSIDs[deviceName]
		if !ok {
			log.Printf("Warning - could not find computer SID for device %s\n", deviceName)
			continue
		}
		computer := internal.Computer{
			ObjectIdentifier: computerSID,
			Owned:            true,
			AlertId:          result["set_AlertId"].(string),
		}
		computers = append(computers, computer)
	}
	metadata := internal.Metadata{
		Type:    internal.DataTypeComputer,
		Methods: 291819,
		Version: 6,
	}
	json, err := FormatBloodHounds2(metadata, computers)
	if err != nil {
		return nil
	}
	fmt.Printf("Uploading changes to %d computers\n", len(computers))
	fmt.Println(string(json))
	return json
}

func FormatBloodHounds(metaData internal.Metadata, sessions []internal.Session) ([]byte, error) {
	payload, err := json.Marshal(sessions)
	if err != nil {
		return nil, err
	}
	jsonData := internal.DataWrapper{
		Metadata: metaData,
		Payload:  payload,
	}
	return json.MarshalIndent(jsonData, "", "  ")
}

func FormatBloodHounds2(metaData internal.Metadata, sessions []internal.Computer) ([]byte, error) {
	payload, err := json.Marshal(sessions)
	if err != nil {
		return nil, err
	}
	jsonData := internal.DataWrapper{
		Metadata: metaData,
		Payload:  payload,
	}
	return json.MarshalIndent(jsonData, "", "  ")
}
