package models

type ClientMessageFormat struct {
	ClientId    int    `json:"clientid"`
	ClientName  string `json:"clientname"`
	Message     string `json:"message"`
	MessageType int    `json:"message_type"`
}

type ServerMessage struct {
	Message     string `json:"message"`
	MessageType int    `json:"message_type"`
	ClientId    int    `json:"clientid"`
}
