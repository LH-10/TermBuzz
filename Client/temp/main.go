package main

import (
	"fmt"

	"github.com/pion/mediadevices"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
	"github.com/pion/webrtc/v4"
)

func main() {

	// fmt.Println("Enter your name :")
	// var new2 stun.ClientAgent

	var newconf webrtc.Configuration
	newconf.ICECandidatePoolSize = 10
	newconf.ICEServers = []webrtc.ICEServer{{URLs: []string{"stun:stun1.l.google.com:19302"}}}

	new, err := webrtc.NewPeerConnection(newconf)
	if err != nil {
		fmt.Print("err:", err, "\n")
		return
	}
	// var mediaconf webrtc.MediaEngine
	// var localaudio mediadevices.Track
	// var msc mediadevices.MediaStreamConstraints
	// var nes mediadevices.AudioSource
	// localaudio = mediadevices.NewAudioTrack(nes, &mediadevices.CodecSelector{})
	// audiotrack, ok := localaudio.(*mediadevices.AudioTrack)
	// if !ok {
	// 	fmt.Println("invlid type")
	// 	return
	// }
	stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(mtc *mediadevices.MediaTrackConstraints) {
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	// audiotrack, ok := stream.GetAudioTracks()[0].(*mediadevices.AudioTrack)
	// if !ok {
	// 	fmt.Println("Wrong type")
	// 	return
	// }
	new.CreateAnswer(nil)
	fmt.Println(stream)
	new.AddTrack(stream.GetAudioTracks()[0])
	sdpMessage, err := new.CreateOffer(nil)
	new.SetLocalDescription(sdpMessage)
	if err != nil {
		fmt.Print("err:", err, "\n")
		return
	}
	new.SetLocalDescription(sdpMessage)
	fmt.Println(sdpMessage)
}
