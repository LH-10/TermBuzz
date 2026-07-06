package models

import "github.com/pion/webrtc/v4"

type ClientMessageFormat struct {
	ClientId     int    `json:"clientid"`
	ClientName   string `json:"clientname"`
	Message      string `json:"message"`
	MessageType  int    `json:"message_type"`
	RecieverName string `json:"reciever_name"`
	webrtc.SessionDescription
}

type ClientMessageFormatFut struct {
	ClientId     int    `json:"clientid"`
	RecieverName string `json:"reciever_name"`
	ClientName   string `json:"clientname"`
	MessageType  int    `json:"message_type"`
	Payload      struct {
		Message string `json:"message"`
		webrtc.SessionDescription
	}
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

