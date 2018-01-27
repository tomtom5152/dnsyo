package dnsyo

import "encoding/json"

// Result contains an answer or error from a single server
type Result struct {
	Answer string
	Error  string `json:",omitempty"`
}

// QueryResults maps servers by name to the results they provide so a more detailed response can be given.
type QueryResults map[string]*Result

// ToJSON prints a verbose JSON representation of the current QueryResults object
func (qr *QueryResults) ToJSON() (string, error) {
	text, err := json.Marshal(qr)
	return string(text), err
}
