package main

import (
	"bufio"
	"context"
	_ "encoding/json"
	"fmt"

	"log"
	"os"
	"time"

	"github.com/LH-10/TermBuzz/Signaling/CMD/constants"
	"github.com/LH-10/TermBuzz/Signaling/CMD/models"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type ClientMessageFormat struct {
	ClientId    int    `json:"clientid"`
	ClientName  string `json:"clientname"`
	Message     string `json:"message"`
	MessageType string `json:"message_type"`
}

func main() {
	fmt.Println("Enter your name :")

	var messageFormat models.ClientMessageFormat
	fmt.Scan(&messageFormat.ClientName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "ws://localhost:8081", nil)
	if err != nil {
		log.Println(err)
	}
	defer c.CloseNow()
	var v models.ServerMessage = models.ServerMessage{}

	fmt.Println("Enter 1 to request cleint ID any other key to skip")
	var choice int
	fmt.Scan(&choice)

	if choice == 1 {
		fmt.Println("requesting Client id.....")
		messageFormat.MessageType = constants.RequestID
		messageFormat.Message = "Requesting ID"
		fmt.Println(messageFormat)
		wsjson.Write(ctx, c, messageFormat)
		fmt.Println("Waiting for id......")
		for {
			wsjson.Read(ctx, c, &v)
			fmt.Println(v)
			if v.MessageType == constants.ClientIDResponse {
				messageFormat.ClientId = v.ClientId
				fmt.Println("Id recieved.")
				break
			}

		}
		fmt.Println(messageFormat)
	}
	messageFormat.MessageType = constants.RequestPeerConnection
	inps := bufio.NewScanner(os.Stdin)
	for inps.Scan() {

		messageFormat.Message = inps.Text()
		if messageFormat.Message == "exit" {
			break
		}
		if messageFormat.Message != "" {
			err = wsjson.Write(ctx, c, messageFormat)

			err = wsjson.Read(ctx, c, &v)
		}
		log.Println(v)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("Enter your message")
	}
	// jsonWithName,err:=json.Marshal(map[string]interface{}{
	// 	"name":
	// })

	for inps.Scan() {

		time.Sleep(time.Second * 1)
	}

	c.Close(websocket.StatusNormalClosure, "closed")
}
