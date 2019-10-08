package vault

import (
	"encoding/json"
	"time"
)

type dataResponse struct {
	Data     map[string]interface{} `json:"data"`
	Metadata struct {
		CreatedTime  time.Time `json:"created_time"`
		DeletionTime string    `json:"deletion_time"`
		Destroyed    bool      `json:"destroyed"`
		Version      int       `json:"version"`
	} `json:"metadata"`
}

func parseDataResponse(data map[string]interface{}) (*dataResponse, error) {
	var resp dataResponse
	return &resp, parseResponse(data, &resp)
}

func parseResponse(data map[string]interface{}, v interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}
