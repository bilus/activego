package activego

type WelcomeResponseTransmission struct {
	Type string `json:"type"`
}

type DisconnectResponseTransmission struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Reconnect bool   `json:"reconnect"`
}

type MessageResponseTransmission struct {
	Message    interface{} `json:"message"`
	Identifier string      `json:"identifier"`
}

type CommandResponseTransmission struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}
