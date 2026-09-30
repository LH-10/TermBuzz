package main

import (
	"fmt"
	"log"

	"github.com/ebitengine/oto/v3"
	opuslib "github.com/pion/opus"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media/samplebuilder"
)

func streamReader(t *webrtc.TrackRemote) {
	var depacktizer rtp.Depacketizer
	switch t.Codec().MimeType {
	case webrtc.MimeTypeOpus:
		depacktizer = &codecs.OpusPacket{}
	default:
		fmt.Print("invalid codec")
	}
	sb := samplebuilder.New(90, depacktizer, t.Codec().ClockRate)
	opusdecoder, err := opuslib.NewDecoderWithOutput(48000, 2)
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
	go func() {
		for {

			packet, _, err := t.ReadRTP()
			if err != nil {
				log.Println(err)
				return
			}

			sb.Push(packet)

			for sample := sb.Pop(); sample != nil; sample = sb.Pop() {
				_, err := rdr.Write(sample.Data)

				if err != nil {
					fmt.Println(err)
				}
			}

		}
	}()
	for {
		if !otoply.IsPlaying() {

			if rdr.HasData() {
				fmt.Println("playing now")
				otoply.Play()
			}
		}
	}

	// time.Sleep(time.Millisecond * 10)
}

func handle_tracks(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
	fmt.Println("codec:", t.Codec())

	fmt.Print("recieving tracks")
	fmt.Print("Paylaod tpe", t.PayloadType())

	go func() {

		streamReader(t)
	}()
}
