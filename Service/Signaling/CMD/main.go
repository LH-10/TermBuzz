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
)

func processClientRequest(clientRequest any) {

}

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
	var clietnMessageFormatHolder models.ClientMessageFormat
	var serverMsg models.ServerMessage
	fmt.Println(clietnMessageFormatHolder)
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
	var MenuForClient [4]string = [4]string{"1.Get Client List\n", "2.Connect to client"}
	// fmt.Print(MenuForClient)
	wsMux := http.NewServeMux()

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
				err = wsjson.Write(ctx, c, models.ServerMessage{Message: clientInfoString.String()})
			case constants.RequestPeerConnection:
				recieverName := clientMessage.RecieverName
				ClientToClient(c, recieverName, ctx, clientMessage.SenderName, clientMessage.Payload.Message)
			case constants.SDPExchange, constants.SDPAnswer:
				recipentData := clients[clientMessage.RecieverName]
				binaryMessage, err := json.Marshal(clientMessage)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println("clientmessage", clientMessage)
				recipentData.conn.Write(ctx, websocket.MessageBinary, binaryMessage)
			// case constants.SDPAnswer:
			// 	recipentData := clients[clientMessage.RecieverName]
			// 	binaryMessage, err := json.Marshal(clientMessage)
			// 	if err != nil {
			// 		fmt.Println("err:", err)
			// 	}
			// 	fmt.Println(clientMessage)
			// 	recipentData.conn.Write(ctx, websocket.MessageBinary, binaryMessage)
			default:
				fmt.Printf("Invalid choicde \n")

			}

			if err != nil {
				fmt.Println(err)
			}

		}

		c.Close(websocket.StatusNormalClosure, "cross origin WebSocket accepted")
	}))

	log.Println("Server Starting on port 8081 ")
	err := http.ListenAndServe("localhost:8081", wsMux)
	if err != nil {
		fmt.Println(err)
	}

}
