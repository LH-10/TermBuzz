package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ebitengine/oto/v3"
	opuslib "github.com/pion/opus"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media/samplebuilder"
)

type AudioController struct {
	ctx    *oto.Context
	Player *oto.Player
	ready  chan struct{}
}

func (ac *AudioController) AssignReader(r io.Reader) {
	<-ac.ready
	ac.Player = ac.ctx.NewPlayer(r)
}

func (ac *AudioController) IsPlayerNil() bool {
	return ac.Player == nil
}

func (ac *AudioController) Initialize() error {
	options := oto.NewContextOptions{SampleRate: 48000, Format: oto.FormatSignedInt16LE, ChannelCount: 2}
	var err error
	ac.ctx, ac.ready, err = oto.NewContext(&options)
	if err != nil {
		fmt.Println("Error", err)
		return err
	}
	return nil
}
func streamReader(t *webrtc.TrackRemote, audio_c *AudioController) {
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
	err = audio_c.Initialize()
	// otoCtxOpts := &oto.NewContextOptions{SampleRate: 48000, Format: oto.FormatSignedInt16LE, ChannelCount: 2}
	// otoCtx, ready, err := oto.NewContext(otoCtxOpts)
	if err != nil {
		fmt.Println(err)
		return
	}
	audio_c.AssignReader(rdr)
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
		if !audio_c.Player.IsPlaying() {

			if rdr.HasData() {
				go audio_c.Player.Play()
				break
			}
		}
	}
	fmt.Println("playing now")
	fmt.Println("IS null:", audio_c.IsPlayerNil())
	for {
		time.Sleep(2 * time.Minute)
	}
}

func handle_tracks(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
	fmt.Println("codec:", t.Codec())

	fmt.Print("recieving tracks")
	fmt.Print("Paylaod tpe", t.PayloadType())

	go func() {
		audio_c := &AudioController{}
		streamReader(t, audio_c)
		audio_c.Player.Play()

		select {}
	}()
}
