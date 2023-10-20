package output_processor

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"falconhound/internal"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type SentinelOutputConfig struct {
	BHQuery          string
	QueryName        string
	QueryDescription string
	QueryEventID     string
}

type SentinelOutputProcessor struct {
	*OutputProcessor
	Config SentinelOutputConfig
}

func (m *SentinelOutputProcessor) BatchSize() int {
	return 0
}

func (m *SentinelOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	jsonData, err := json.Marshal(QueryResults)
	if err != nil {
		return err
	}
	sentinelData := map[string]interface{}{
		"Name":        m.Config.QueryName,
		"Description": m.Config.QueryDescription,
		"EventID":     m.Config.QueryEventID,
		"BHQuery":     m.Config.BHQuery,
		"EventData":   string(jsonData),
	}

	customerId := m.Credentials.SentinelWorkspaceID
	sharedKey := m.Credentials.SentinelSharedKey
	logName := m.Credentials.SentinelTargetTable
	timeStampField := "DateValue"

	data, err := json.Marshal(sentinelData)
	if err != nil {
		return err
	}

	Senddata := string(data)

	dateString := time.Now().UTC().Format(time.RFC1123)
	dateString = strings.Replace(dateString, "UTC", "GMT", -1)

	stringToHash := "POST\n" + strconv.Itoa(utf8.RuneCountInString(Senddata)) + "\napplication/json\n" + "x-ms-date:" + dateString + "\n/api/logs"
	hashedString, err := SentinelBuildSignature(stringToHash, sharedKey)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	signature := "SharedKey " + customerId + ":" + hashedString
	url := "https://" + customerId + ".ods.opinsights.azure.com/api/logs?api-version=2016-04-01"

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(data)))
	if err != nil {
		return err
	}

	req.Header.Add("Log-Type", logName)
	req.Header.Add("Authorization", signature)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-ms-date", dateString)
	req.Header.Add("time-generated-field", timeStampField)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending data to Sentinel: ", err.Error())
		return err
	}
	return resp.Body.Close()
}

func SentinelBuildSignature(message, secret string) (string, error) {

	keyBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}
