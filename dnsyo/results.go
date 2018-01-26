package dnsyo

import "encoding/json"

type Result struct {
	Answer string
	Error  string `json:",omitempty"`
}

type QueryResults map[string]*Result

func (qr *QueryResults) ToJson() (string, error) {
	text, err := json.Marshal(qr)
	return string(text), err
}
