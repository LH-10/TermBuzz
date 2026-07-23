package main

import (
	"bufio"
	"context"
	"encoding/json"
	_ "encoding/json"
	"fmt"
	"runtime"

	"log"
	"os"

	"github.com/LH-10/TermBuzz/shared/constants"
	"github.com/LH-10/TermBuzz/shared/models"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/pion/interceptor"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
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

// var peerConn *webrtc.PeerConnection

// var clientGlob = struct {
// 	peerConn *webrtc.PeerConnection
// }{}

func getAudio(codecSelector *mediadevices.CodecSelector) ([]mediadevices.Track, error) {
	stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(mtc *mediadevices.MediaTrackConstraints) {
			// mtc.AudioConstraints=prop.AudioConstraints{}
		},
		Codec: codecSelector,
	})

	if err != nil {
		return nil, err
	}

	audioTrack := stream.GetAudioTracks()
	return audioTrack, nil
}

func createPeerConn() (*webrtc.PeerConnection, error) {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		panic(err)
	}
	opusParams, err := opus.NewParams()
	if err != nil {
		fmt.Println("opus err")
		panic(err)
	}
	codecSelector := mediadevices.NewCodecSelector(mediadevices.WithAudioEncoders(&opusParams))
	codecSelector.Populate(m)
	i := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		panic(err)
	}
	var config webrtc.Configuration
	config.ICECandidatePoolSize = 1
	config.ICEServers = []webrtc.ICEServer{{URLs: []string{"stun:stun1.l.google.com:19302"}}}
	peerConn, err := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i)).NewPeerConnection(config)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	audioTrack, err := getAudio(codecSelector)
	if err != nil {
		return nil, err
	}
	for _, track := range audioTrack {

		peerConn.AddTrack((track))
	}

	return peerConn, nil
}

func handleMessage(peerConn *webrtc.PeerConnection, messenger *messaging, messageToServer *models.ClientMessageFormatFut, inps *bufio.Scanner) {
	switch messageToServer.Payload.Message {
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
		messageToServer.Payload.Message = inps.Text()
	case "3":
		fmt.Println("Enter reciver name:")
		inps.Scan()
		recvr := inps.Text()
		fmt.Println("making call")
		makeCall(peerConn, messenger, recvr)
		fmt.Println("made call")
	default:
		fmt.Println("Invlaid choice")

	}

}

var localaudio mediadevices.AudioTrack

func makeCall(peerConn *webrtc.PeerConnection, messenger *messaging, reciever string) error {
	if peerConn == nil {
		var err error
		peerConn, err = createPeerConn()
		if err != nil {
			return err
		}
	}
	// peerConn.OnICECandidate(func(i *webrtc.ICECandidate) {
	// 	messenger.send(models.ClientMessageFormat{
	// 		RecieverName: reciever,
	// 		MessageType:  2300, //candidtae type new
	// 		Message:      i.String(),
	// 	})
	// })
	sdp, err := peerConn.CreateOffer(&webrtc.OfferOptions{})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = peerConn.SetLocalDescription(sdp)
	if err != nil {
		return err
	}
	messenger.send(models.ClientMessageFormatFut{

		RecieverName: reciever,
		MessageType:  constants.SDPExchange,
		Payload: struct {
			Message string `json:"message"`
			*webrtc.SessionDescription
		}{
			Message:            "SDP",
			SessionDescription: &sdp,
		},
	})
	return nil
}

func incomingCall(peerConn *webrtc.PeerConnection, remoteSDP *webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	var err error
	if peerConn == nil {

		peerConn, err = createPeerConn()
		if err != nil {
			fmt.Println("error while creating peerconn obj")
			return webrtc.SessionDescription{}, err
		}
	}
	fmt.Println("sdp from remote peer :", *remoteSDP, "\n\n ")
	peerConn.SetRemoteDescription(*remoteSDP)
	ansSDP, err := peerConn.CreateAnswer(nil)
	if err != nil {
		fmt.Println(err)
		return webrtc.SessionDescription{}, err
	}
	fmt.Println("created ansSPD:", ansSDP)
	err = peerConn.SetLocalDescription(ansSDP)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}
	fmt.Println("here")
	return ansSDP, nil
}

func readAnswer(peerConn *webrtc.PeerConnection, remoteSDP webrtc.SessionDescription) error {
	err := peerConn.SetRemoteDescription(remoteSDP)
	return err
}

type messaging struct {
	conn *websocket.Conn
	ctx  context.Context
}

func (msg *messaging) send(message models.ClientMessageFormatFut) {

	binMsg, err := SerializeMessage(message)

	if err != nil {
		log.Println("Error during serialize", err)
		return
	}
	msg.conn.Write(msg.ctx, websocket.MessageBinary, binMsg)
}

func (msg *messaging) read() {

}

func main() {
	fmt.Println("Enter your name :")

	var err error
	var peerConn *webrtc.PeerConnection
	peerConn, err = createPeerConn()
	if err != nil {
		log.Println(err)
	}

	var messageToServer models.ClientMessageFormatFut
	fmt.Scan(&messageToServer.SenderName)
	msgr := &messaging{}
	msgr.ctx = context.Background()
	ipadd := "127.0.0.1"
	port := "8081"
	address := fmt.Sprintf("ws://%s:%s", ipadd, port)
	msgr.conn, _, err = websocket.Dial(msgr.ctx, address, nil)
	if err != nil {
		log.Println(err)
	}
	defer msgr.conn.CloseNow()
	var v models.ServerMessage

	fmt.Println("Enter 1 to request cleint ID any other key to skip")
	var choice int
	fmt.Scan(&choice)

	if choice == 1 {
		fmt.Println("requesting Client id.....")
		messageToServer.MessageType = constants.RequestID
		messageToServer.Payload.Message = "REq"
		fmt.Println(messageToServer)
		tempbt, err := json.Marshal(messageToServer)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(tempbt, &messageToServer)
		if err != nil {
			fmt.Println("err:", err, err.Error())
			fmt.Println("here")
		}

		err = wsjson.Write(msgr.ctx, msgr.conn, messageToServer)
		if err != nil {
			fmt.Println("err:", err)
			return
		}
		fmt.Println("Waiting for id......")
		for {
			err = wsjson.Read(msgr.ctx, msgr.conn, &v)
			if err != nil {
				fmt.Println("Error:", err)
				runtime.Goexit()
			}
			fmt.Println(v)
			if v.MessageType == constants.ClientIDResponse {
				messageToServer.ClientId = v.ClientId
				fmt.Println("Id recieved.")
				break
			}

		}
		fmt.Println(messageToServer.SenderName, " ", messageToServer.ClientId)
	}

	messageToServer.MessageType = constants.Blank
	inps := bufio.NewScanner(os.Stdin)
	err = wsjson.Read(msgr.ctx, msgr.conn, &v)
	menu := v.Message
	var newReciever models.ClientMessageFormatFut

	go func() {
		for {
			err = wsjson.Read(msgr.ctx, msgr.conn, &newReciever)
			if err != nil {
				fmt.Println(err.Error(), err)
				if msgr.conn.Ping(msgr.ctx) != nil {
					fmt.Println("Cannot connect")
					runtime.Goexit()
				}
			}
			fmt.Println("Message of type", newReciever.MessageType)
			fmt.Println(newReciever.Payload.Message, "\t\n\n", newReciever)
			switch newReciever.MessageType {
			case constants.SDPExchange:
				fmt.Println("SDP Exchange initiated")
				ans, err := incomingCall(peerConn, newReciever.Payload.SessionDescription)
				if err != nil {
					fmt.Println(err)
					continue
				}
				messageToServer.RecieverName = newReciever.SenderName
				messageToServer.SenderName = newReciever.RecieverName
				messageToServer.MessageType = constants.SDPAnswer
				messageToServer.Payload.Message = "SDPAnswer"
				messageToServer.Payload.SessionDescription = &ans
				msgr.send(messageToServer)
				fmt.Println("SDP answer sent")
			case constants.SDPAnswer:
				fmt.Println("Got an answer")
				err := readAnswer(peerConn, *newReciever.Payload.SessionDescription)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println("Got an answer", *peerConn.CurrentRemoteDescription() == *newReciever.Payload.SessionDescription)
			}
		}
	}()

	for inps.Scan() {
		messageToServer.Payload.Message = inps.Text()

		if messageToServer.Payload.Message == "exit" {
			fmt.Println("Exiting")
			break
		}

		handleMessage(peerConn, msgr, &messageToServer, inps)

		if messageToServer.Payload.Message != "" {
			err = wsjson.Write(msgr.ctx, msgr.conn, messageToServer)
			messageToServer.Payload.Message = ""

		}

		// log.Println(v)
		if err != nil {
			log.Println(err)
			if msgr.conn.Ping(msgr.ctx) != nil {
				log.Println("Cannot connect")
				msgr.conn.CloseNow()
				runtime.Goexit()
			}
		}

		fmt.Println(menu)
	}

	msgr.conn.Close(websocket.StatusNormalClosure, "closed")
}
