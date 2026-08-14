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
	"bufio"
	"context"
	"encoding/json"
	_ "encoding/json"
	"errors"
	"flag"
	"fmt"
	"runtime"

	"log"
	"os"

	"github.com/LH-10/TermBuzz/shared/constants"
	"github.com/LH-10/TermBuzz/shared/models"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/ebitengine/oto/v3"
	"github.com/pion/interceptor"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
	piopus "github.com/pion/opus"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media/samplebuilder"
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

type opusCodecReader struct {
	buf         []byte
	opusDecoder piopus.Decoder
	bufOffset   int
	buffer      [10][]byte
	indx        int
	readindx    int
}

func (ocr *opusCodecReader) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	var out []byte = make([]byte, 3890) //2*1920+50
	_, _, err = ocr.opusDecoder.Decode(p, out)
	// fmt.Println("band:", band, "\nstero?", isStr)
	if err != nil {
		fmt.Println("here")
		fmt.Println("err:", err)
		return 0, err
	}

	copy(ocr.buffer[ocr.indx], out)
	ocr.indx = (ocr.indx + 1) % len(ocr.buffer)
	return len(out), nil

}
func (ocr *opusCodecReader) HasData() bool {
	return ocr.indx > 2
}

func (ocr *opusCodecReader) Read(p []byte) (n int, err error) {
	// n = copy(p, ocr.buf)
	for ocr.readindx == ocr.indx {
	}
	n = copy(p, ocr.buffer[ocr.readindx])
	// fmt.Println("Read called with", n)
	ocr.readindx = (ocr.readindx + 1) % len(ocr.buffer)
	// if n == 0 {
	// 	return n, io.EOF
	// }
	return n, err

}

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

// creates peerconn with codecs and registers events  such as onTrack
func createPeerConn() (*webrtc.PeerConnection, error) {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		panic(err)
	}
	opusParams, err := opus.NewParams()
	opusParams.RTPCodec().ClockRate = 48000
	opusParams.RTPCodec().Channels = 0
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
	peerConn.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) { fmt.Print("webrtc Connection", pcs, "\n") })
	peerConn.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) { fmt.Print("IceConnection State:", is.String()) })
	peerConn.OnICEGatheringStateChange(func(is webrtc.ICEGatheringState) {
		fmt.Print("Ice Gethering state:", is, "\n\n")
	})

	peerConn.OnTrack(func(t *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		fmt.Println("codec:", t.Codec())

		fmt.Print("recieving tracks")
		fmt.Print("Paylaod tpe", t.PayloadType())
		readStream := func(t *webrtc.TrackRemote) {
			var depacktizer rtp.Depacketizer
			switch t.Codec().MimeType {
			case webrtc.MimeTypeOpus:
				depacktizer = &codecs.OpusPacket{}
			default:
				fmt.Print("invalid codec")
			}

			sb := samplebuilder.New(90, depacktizer, t.Codec().ClockRate)
			opusdecoder, err := piopus.NewDecoderWithOutput(48000, 2)
			if err != nil {
				fmt.Println(err)
				return
			}
			var dt [10][]byte
			for i := range dt {
				dt[i] = make([]byte, 3980)
			}
			rdr := &opusCodecReader{opusDecoder: opusdecoder, buf: make([]byte, 0), buffer: dt, indx: 0, readindx: 0}
			otoCtxOpts := &oto.NewContextOptions{SampleRate: 48000, Format: oto.FormatSignedInt16LE, ChannelCount: 2}
			otoCtx, ready, err := oto.NewContext(otoCtxOpts)
			if err != nil {
				fmt.Println(err)
				return
			}
			<-ready
			otoply := otoCtx.NewPlayer(rdr)
			thirdRead := 0
			// gotSample := make(chan struct{})
			go func() {
				for {
					if thirdRead%3 == 0 {
						// fmt.Println("reading.....")
					}

					packet, _, err := t.ReadRTP()
					if err != nil {
						log.Println(err)
						return
					}

					sb.Push(packet)

					for sample := sb.Pop(); sample != nil; sample = sb.Pop() {
						_, err := rdr.Write(sample.Data)

						if thirdRead%3 == 0 {
							// fmt.Println("extract sample on 3rd")
							// fmt.Println("wrote n:", n)
						}
						if err != nil {
							fmt.Println(err)
						}
						// go func(){gotSample <- struct{}{}
					}

					thirdRead++

				}
			}()
			for {
				// <-gotSample
				if !otoply.IsPlaying() {

					if rdr.HasData() {
						fmt.Println("playing now")
						otoply.Play()
					}
				}
			}

			// time.Sleep(time.Millisecond * 10)
		}
		go func() {

			readStream(t)
		}()
	})
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

// initiate sdp exchange
func makeCall(peerConn *webrtc.PeerConnection, messenger *messaging, reciever string) error {
	if peerConn == nil {
		var err error
		peerConn, err = createPeerConn()
		if err != nil {
			return err
		}
	}
	peerConn.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			fmt.Println("caller has nil candidate")
			return
		}
		fmt.Println("on ice candidte fired")
		candidate := i.ToJSON()
		messenger.send(models.ClientMessageFormatFut{
			RecieverName: reciever,
			MessageType:  constants.Candidate, //candidtae type new
			Payload: struct {
				Message string "json:\"message\""
				*webrtc.SessionDescription
				*webrtc.ICECandidateInit
			}{
				Message:          i.String(),
				ICECandidateInit: &candidate,
			},
		})
	})
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
			*webrtc.ICECandidateInit
		}{
			Message:            "SDP",
			SessionDescription: &sdp,
			ICECandidateInit:   nil,
		},
	})
	return nil
}

// read remote sdp and generate answer
func incomingCall(reciever string, messenger messaging, peerConn *webrtc.PeerConnection, remoteSDP *webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	var err error
	if peerConn == nil {

		peerConn, err = createPeerConn()
		if err != nil {
			fmt.Println("error while creating peerconn obj")
			return webrtc.SessionDescription{}, err
		}
	}
	peerConn.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			fmt.Println("nil  candidate object")
			return
		}
		fmt.Println("on ice candidte fired")
		candidate := i.ToJSON()
		messenger.send(models.ClientMessageFormatFut{
			RecieverName: reciever,
			MessageType:  constants.Candidate, //candidtae type new
			Payload: struct {
				Message string "json:\"message\""
				*webrtc.SessionDescription
				*webrtc.ICECandidateInit
			}{
				Message:          i.String(),
				ICECandidateInit: &candidate,
			},
		})
	})
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
		panic(fmt.Sprintf("Error during serialize %v", err))

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
	ipadd := flag.String("ip", "192.168.1.5", "ipaddress of server")
	port := "8081"
	flag.Parse()
	address := fmt.Sprintf("ws://%s:%s", *ipadd, port)
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
			case constants.SDPExchange: //someone sent sdp to connect
				fmt.Println("SDP Exchange initiated")
				ans, err := incomingCall(newReciever.SenderName, *msgr, peerConn, newReciever.Payload.SessionDescription)
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
			case constants.SDPAnswer: // got sdp answer as response from peer
				fmt.Println("Got an answer")
				err := readAnswer(peerConn, *newReciever.Payload.SessionDescription)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println("Got an answer", *peerConn.CurrentRemoteDescription() == *newReciever.Payload.SessionDescription)
			case constants.Candidate:
				fmt.Print("in ice candidates")
				if newReciever.Payload.ICECandidateInit == nil {
					fmt.Println(errors.New("Empty candidate in Paylod"))
				}
				fmt.Println("all set adding candidates")
				err := peerConn.AddICECandidate(*newReciever.Payload.ICECandidateInit)
				if err != nil {
					log.Println(err)
				} else {
					fmt.Println("Added ", *newReciever.Payload.ICECandidateInit, " :CANDIDATE\n")
				}
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
