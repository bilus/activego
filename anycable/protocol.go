package anycable

type MessageResponseTransmission struct {
	Message    interface{} `json:"message"`
	Identifier string      `json:"identifier"`
}

type CommandResponseTransmission struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}
