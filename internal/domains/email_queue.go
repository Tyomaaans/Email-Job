package domains

import (
	"encoding/json"
)

type EmailType string

const (
	EmailTypeContact     EmailType = "contact_me"
	EmailTypeDemoRequest EmailType = "demo_request"
)

type EmailQueueEntity struct {
	ID        string          `json:"id"`
	IpAddress string          `json:"ip_address"`
	Name      string          `json:"name"`
	Project   *Project        `json:"project"`
	Attempt   int             `json:"attempt"`
	Type      EmailType       `json:"type"`
	To        string          `json:"to"`
	Subject   string          `json:"subject"`
	Payload   json.RawMessage `json:"payload"`
}

type ContactQueueEntity struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type DemoRequestQueueEntity struct {
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Project Project `json:"project"`
	Message string  `json:"message"`
}