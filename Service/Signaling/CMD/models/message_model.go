package models

type ClientMessageFormat struct {
	ClientId     int    `json:"clientid"`
	ClientName   string `json:"clientname"`
	Message      string `json:"message"`
	MessageType  int    `json:"message_type"`
	RecieverName string `json:"reciever_name"`
}

type ServerMessage struct {
	Message     string `json:"message"`
	MessageType int    `json:"message_type"`
	ClientId    int    `json:"clientid"`
}

type Message interface {
	getMessage() string
	getMessageType() int
}
