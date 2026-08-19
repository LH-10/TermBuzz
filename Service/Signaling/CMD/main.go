// SPDX-License-Identifier: GPL-2.0

//TermBuzz
// Copyright (C) 2026  Lalit Hinduja

//    This program is free software; you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation;

//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/LH-10/TermBuzz/shared/constants"
	"github.com/LH-10/TermBuzz/shared/models"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/pion/webrtc/v4"
)

type Client struct {
	conn *websocket.Conn
	id   int
	name string
}

type clientName = string

var clients map[clientName]*Client = make(map[string]*Client)

func ClientToClient(sender *websocket.Conn, recievername string, ctx context.Context, sendername string, message string) {
	senderClient, ok := clients[sendername]
	if !ok {
		fmt.Println("Sender does not exists")
		return
	}
	if senderClient.conn != sender {
		fmt.Println(sendername, " has wrong connection id ")
		return
	}
	var serverMsg models.ServerMessage
	var recieverConnection *websocket.Conn
	recieverConnection = clients[recievername].conn
	fmt.Println("Sender", sendername, "\n reciever conn", recieverConnection)
	message = sendername + ":" + message
	serverMsg.Message = message
	serverMsg.MessageType = constants.PeerChat

	wsjson.Write(ctx, recieverConnection, serverMsg)
}

var numclient int

func main() {
	var MenuForClient [4]string = [4]string{"1.Get Client List\n", "2.Connect to client , 3.SDP exchange"}
	// fmt.Print(MenuForClient)
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("works"))
	})
	wsMux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})
		currentClient := &Client{conn: c, id: int(time.Now().Unix())}

		// clients = append(clients, currentClient)
		fmt.Println("New connection", clients)
		if err != nil {
			log.Println(err)
		}
		defer c.CloseNow()
		numclient++
		ctx := context.Background()
		// var v any
		var clientMessage models.ClientMessageFormatFut
		err = wsjson.Read(ctx, c, &clientMessage) //Read is a blocking operation

		if err != nil {
			log.Println(err)
			return
		}
		(currentClient).name = clientMessage.SenderName
		clients[currentClient.name] = currentClient
		if clientMessage.MessageType == constants.RequestID {
			err = wsjson.Write(ctx, c, models.ServerMessage{MessageType: constants.ClientIDResponse, ClientId: currentClient.id, Message: "Your Id"})
			if err != nil {
				fmt.Println(err)

			}
			fmt.Println("ID Sent")
		}
		fmt.Printf("connections ")
		for i := range clients {
			fmt.Println(clients[i])
		}

		err = wsjson.Write(ctx, c, models.ServerMessage{Message: fmt.Sprintf("%v", MenuForClient)})
		if err != nil {
			log.Println(err)
		}

		for {

			err = wsjson.Read(ctx, c, &clientMessage)

			if err != nil {
				log.Println(err)
				break
			}
			log.Printf("recieved: %v", clientMessage)

			switch clientMessage.MessageType {

			case constants.RequestPeerList:
				var clientInfoString strings.Builder
				for i := range clients {
					clientInfoString.WriteString(fmt.Sprintf("%v", *clients[i]))
				}
				err = wsjson.Write(ctx, c, models.ClientMessageFormatFut{Payload: struct {
					Message string `json:"message"`
					*webrtc.SessionDescription
					*webrtc.ICECandidateInit
				}{Message: clientInfoString.String()}})
			case constants.RequestPeerConnection:
				recieverName := clientMessage.RecieverName
				ClientToClient(c, recieverName, ctx, clientMessage.SenderName, clientMessage.Payload.Message)
			case constants.SDPExchange, constants.SDPAnswer:
				if clientMessage.RecieverName == "" {
					fmt.Println("empty Reciver field")
				}
				clientMessage.SenderName = currentClient.name
				recipentData := clients[clientMessage.RecieverName]
				var binaryMessage []byte
				binaryMessage, err = json.Marshal(clientMessage)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println(clientMessage.SenderName, "->", clientMessage.RecieverName)
				fmt.Println("clientmessage", clientMessage)
				fmt.Println("Before wirte called")
				err = recipentData.conn.Write(ctx, websocket.MessageBinary, binaryMessage)
			case constants.Candidate:
				fmt.Println("\n\nCANDIDATE : ********", "from ", currentClient.name, "\nto ", clientMessage.RecieverName, "*******", "\n\nPayload:", clientMessage.Payload)
				if clientMessage.RecieverName == "" {
					fmt.Println("empty Reciver field")
					return
				}
				clientMessage.SenderName = currentClient.name
				recipentData := clients[clientMessage.RecieverName]
				var binaryMessage []byte
				binaryMessage, err = json.Marshal(clientMessage)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println(clientMessage.SenderName, "->", clientMessage.RecieverName)
				fmt.Println("clientmessage", clientMessage)
				fmt.Println("Before wirte called")
				err = recipentData.conn.Write(ctx, websocket.MessageBinary, binaryMessage)

			// case constants.SDPAnswer:
			// 	recipentData := clients[clientMessage.RecieverName]
			// 	binaryMessage, err := json.Marshal(clientMessage)
			// 	if err != nil {
			// 		fmt.Println("err:", err)
			// 	}
			// 	fmt.Println(clientMessage)
			// 	recipentData.conn.Write(ctx, websocket.MessageBinary, binaryMessage)
			default:
				fmt.Println("Message of type:", clientMessage.MessageType, "is:")
				fmt.Printf("Invalid choicde \n")

			}

			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("After write ")

		}

		c.Close(websocket.StatusNormalClosure, "cross origin WebSocket accepted")
	}))

	log.Println("Server Starting on port 8081 ")
	err := http.ListenAndServe(":8081", wsMux)
	if err != nil {
		fmt.Println(err)
	}

}
