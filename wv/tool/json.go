package tool

import "encoding/json"

type JSData struct {
	Option string `json:"option"`
	Type   string `json:"type"`
	Data   any    `json:"data"`
}

func ToJSData(data string) *JSData {
	var m JSData
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil
	}
	return &m
}
