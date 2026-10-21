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

func (ac *AudioController) StopAudio() error {
	return ac.ctx.Suspend()
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
	rtp_read_stop := make(chan error)
	go func() {
		for {
			t.SetReadDeadline(time.Now().Add(10 * time.Second))

			packet, _, err := t.ReadRTP()
			if err != nil {
				log.Println(err)
				rtp_read_stop <- err
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
		fmt.Println("rtp error", <-rtp_read_stop)
		rdr.stopReading = true
		audio_c.Player.Pause()
		audio_c.Player = nil
		rdr = nil
		break
	}
}

func handle_tracks(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
	fmt.Println("codec:", t.Codec())

	fmt.Print("recieving tracks")
	fmt.Print("Paylaod tpe", t.PayloadType())

	go func() {
		audio_c := &AudioController{}
		streamReader(t, audio_c)
		err := audio_c.StopAudio()
		if err != nil {
			fmt.Println("Error while trying to suspend audio", err)
		}

	}()
}
