package main

import (
	"bufio"
	"context"
	_ "encoding/json"
	"fmt"
	"runtime"

	"log"
	"os"

	"github.com/LH-10/TermBuzz/Client/constants"
	"github.com/LH-10/TermBuzz/Client/models"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/pion/mediadevices"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
	"github.com/pion/stun"
	"github.com/pion/webrtc/v4"
)

// func chatWithClient(ctx context.Context, conn *websocket.Conn, messageStructure models.ClientMessageFormat) {
// 	inps := bufio.NewScanner(os.Stdin)
// 	inps.Scan()
// 	fmt.Println("You are now connected with other client")
// 	messageStructure.MessageType = constants.PeerChat
// 	PeerMessage := new(models.ClientMessageFormat)

// 	go func() {
// 		for {
// 			wsjson.Read(ctx, conn, PeerMessage)
// 			if PeerMessage.Message == "exit" {
// 				fmt.Println("exiting chat")
// 				break
// 			}
// 			fmt.Println(PeerMessage.ClientName, ": ", PeerMessage.Message)
// 		}
// 	}()
// 	for inps.Scan() {

// 		messageStructure.Message = inps.Text()
// 		if messageStructure.Message == "exit" {
// 			wsjson.Write(ctx, conn, messageStructure)
// 			fmt.Println("Ending chat")
// 			break
// 		}
// 		wsjson.Write(ctx, conn, messageStructure)

// 	}
// }

func createPeerConn() (*webrtc.PeerConnection, error) {
	var config webrtc.Configuration
	config.ICECandidatePoolSize = 10
	config.ICEServers = []webrtc.ICEServer{{URLs: []string{"stun:stun1.l.google.com:19302"}}}
	peerConn, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return peerConn, nil
}

func handleMessage(messageToServer *models.ClientMessageFormat, inps *bufio.Scanner) {
	switch messageToServer.Message {
	case "1":
		messageToServer.MessageType = constants.RequestPeerList
	case "2":
		messageToServer.MessageType = constants.RequestPeerConnection
		clientName := ""
		fmt.Println("Enter Client name to connect with")
		inps.Scan()
		clientName = inps.Text()
		messageToServer.RecieverName = clientName
		fmt.Print("Enter Message:")
		inps.Scan()
		messageToServer.Message = inps.Text()
	default:
		fmt.Println("Invlaid choice")

	}

}

var new2 stun.ClientAgent
var localaudio mediadevices.AudioTrack

func makeCall(peerConn *webrtc.PeerConnection, messenger *messanging, reciever string) {
	sdp, err := peerConn.CreateOffer(&webrtc.OfferOptions{})
	if err != nil {
		fmt.Println(err)
	}
	peerConn.SetLocalDescription(sdp)
	messenger.send(models.ClientMessageFormat{
		RecieverName:       reciever,
		SessionDescription: sdp,
		Message:            "SDP",
		MessageType:        constants.SDPExchange,
	})
}

type messanging struct {
	conn *websocket.Conn
	ctx  context.Context
}

func (msg *messanging) send(message models.ClientMessageFormat) {

	binMsg, err := SerializeMessage(message)

	if err != nil {
		log.Println("Error during serialize", err)
	}
	msg.conn.Write(msg.ctx, websocket.MessageBinary, binMsg)
}

func (msg *messanging) read() {

}

func main() {
	fmt.Println("Enter your name :")

	stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(mtc *mediadevices.MediaTrackConstraints) {
			// mtc.AudioConstraints=prop.AudioConstraints{}
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(localaudio)
	peerConn, err := createPeerConn()
	if err != nil {
		log.Println(err)
	}
	peerConn.AddTrack(stream.GetAudioTracks()[0])
	var msc mediadevices.MediaStreamConstraints
	sdpMessage, err := peerConn.CreateOffer(&webrtc.OfferOptions{})
	if err != nil {
		fmt.Print("err:", err, "\n")
		return
	}
	peerConn.SetLocalDescription(sdpMessage)
	fmt.Println("SDP:", sdpMessage, "MSC", msc)
	fmt.Println(new2)
	var messageToServer models.ClientMessageFormat
	fmt.Scan(&messageToServer.ClientName)

	ctx := context.Background()

	c, _, err := websocket.Dial(ctx, "ws://localhost:8081", nil)
	if err != nil {
		log.Println(err)
	}
	defer c.CloseNow()
	var v models.ServerMessage

	fmt.Println("Enter 1 to request cleint ID any other key to skip")
	var choice int
	fmt.Scan(&choice)

	if choice == 1 {
		fmt.Println("requesting Client id.....")
		messageToServer.MessageType = constants.RequestID
		fmt.Println(messageToServer)
		err = wsjson.Write(ctx, c, messageToServer)
		fmt.Println("Waiting for id......")
		for {
			err = wsjson.Read(ctx, c, &v)
			if err != nil {
				fmt.Println("85 ", err)
			}
			fmt.Println(v)
			if v.MessageType == constants.ClientIDResponse {
				messageToServer.ClientId = v.ClientId
				fmt.Println("Id recieved.")
				break
			}

		}
		fmt.Println(messageToServer.ClientName, " ", messageToServer.ClientId)
	}

	messageToServer.MessageType = constants.Blank
	inps := bufio.NewScanner(os.Stdin)
	err = wsjson.Read(ctx, c, &v)
	menu := v.Message
	var newReciever models.ServerMessage

	go func() {
		for {
			err = wsjson.Read(ctx, c, &newReciever)
			if err != nil {
				fmt.Println(err.Error(), "111")
				if c.Ping(ctx) != nil {
					fmt.Println("Cannot connect")
					runtime.Goexit()
				}
			}
			fmt.Println("Message of type", newReciever.MessageType)
			fmt.Println(newReciever.Message, "\n\n\t", newReciever)
		}
	}()

	for inps.Scan() {
		messageToServer.Message = inps.Text()

		if messageToServer.Message == "exit" {
			fmt.Println("Exiting")
			break
		}

		handleMessage(&messageToServer, inps)

		if messageToServer.Message != "" {
			err = wsjson.Write(ctx, c, messageToServer)
			messageToServer.Message = ""

		}

		// log.Println(v)
		if err != nil {
			log.Println(err)
			if c.Ping(ctx) != nil {
				log.Println("Cannot connect")
				c.CloseNow()
				runtime.Goexit()
			}
		}

		fmt.Println(menu)
	}

	c.Close(websocket.StatusNormalClosure, "closed")
}
