package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/LH-10/TermBuzz/Signaling/CMD/constants"
	"github.com/LH-10/TermBuzz/Signaling/CMD/models"
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

var clients []*Client

func ClientToClient(sender *websocket.Conn, recievername string, ctx context.Context, message string) {
	var sendername string
	var clietnMessageFormatHolder models.ClientMessageFormat
	fmt.Println(clietnMessageFormatHolder)
	var reciever *websocket.Conn
	for i := range clients {
		if sender == clients[i].conn {
			sendername = clients[i].name
			continue
		}
		if recievername == clients[i].name {
			reciever = clients[i].conn
		}
	}
	fmt.Println("Sender", sendername, "\n reciever conn", reciever)
	message += sendername
	wsjson.Write(ctx, reciever, message)
}

var numclient int

func main() {
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})
		currentClient := &Client{conn: c, id: int(time.Now().Unix())}
		clients = append(clients, currentClient)
		fmt.Println("New connection", clients)
		if err != nil {
			log.Println(err)
		}
		defer c.CloseNow()
		numclient++
		ctx := context.Background()
		var v any
		var clientMessage models.ClientMessageFormat
		fmt.Println("Woke up")
		err = wsjson.Read(ctx, c, &clientMessage)

		if err != nil {
			log.Println(err)
			return
		}
		(currentClient).name = clientMessage.ClientName
		fmt.Println(currentClient.name)
		if clientMessage.MessageType == constants.RequestID {
			err = wsjson.Write(ctx, c, models.ServerMessage{MessageType: constants.ClientIDResponse, ClientId: currentClient.id, Message: "Your Id"})
			if err != nil {
				fmt.Println(err)

			}
			fmt.Println("ID Sent")
		}
		fmt.Printf("connections ")
		for _, values := range clients {
			fmt.Println(*values)
		}
		fmt.Println(v, "\b\b\b")

		for {

			// go func() {
			// 	if numclient != len(clients) {
			// 		fmt.Println("here", numclient, len(clients))
			// 		wsjson.Write(ctx, c, fmt.Sprintf("others joined%v", clients))
			// 		numclient = len(clients)
			// 	}
			// }()
			err = wsjson.Read(ctx, c, &clientMessage)

			if err != nil {
				log.Println(err)
				log.Println("here")
				break
			}
			log.Printf("recieved: %v", clientMessage)

			if clientMessage.MessageType != constants.RequestID {
				wsjson.Write(ctx, c, models.ServerMessage{Message: "Hi i recieved your message"})
			}
		}

		c.Close(websocket.StatusNormalClosure, "cross origin WebSocket accepted")
	}))

	err := http.ListenAndServe("localhost:8081", nil)
	if err != nil {
		fmt.Println(err)
	}
	log.Println("Server Started ")
	// fmt.Println("Hello",wsfn)

}
