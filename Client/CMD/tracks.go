package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/jitterbuffer"
	opuslib "github.com/pion/opus"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
)

var oto_ctx *oto.Context = nil

type AudioController struct {
	ctx    *oto.Context
	Player *oto.Player
	ready  chan struct{}
}

func (ac *AudioController) StopAudio() error {
	return ac.ctx.Suspend()
}
func (ac *AudioController) AssignReader(r io.Reader) {
	ac.Player = ac.ctx.NewPlayer(r)
}

func (ac *AudioController) IsPlayerNil() bool {
	return ac.Player == nil
}

func (ac *AudioController) Initialize() error {
	if oto_ctx != nil {
		ac.ctx = oto_ctx
		ac.ctx.Resume()
		return ac.ctx.Err()
	}
	options := oto.NewContextOptions{SampleRate: 48000, Format: oto.FormatSignedInt16LE, ChannelCount: 2}
	var err error
	oto_ctx, ac.ready, err = oto.NewContext(&options)
	if err != nil {
		fmt.Println("Error", err)
		return err
	}
	ac.ctx = oto_ctx
	<-ac.ready
	if err := ac.ctx.Err(); err != nil {
		return err
	}
	return nil
}
func streamReader(t *webrtc.TrackRemote, rtpreader interceptor.RTPReader) {
	var depacktizer rtp.Depacketizer
	switch t.Codec().MimeType {
	case webrtc.MimeTypeOpus:
		depacktizer = &codecs.OpusPacket{}
		_ = depacktizer //was using with sample builder but since i removed sample builder this might not be needed
		fmt.Println("Opus Code")
	default:
		fmt.Println("invalid codec")
		return
	}
	audio_c := AudioController{}
	// sb := samplebuilder.New(90, depacktizer, t.Codec().ClockRate)
	opusdecoder, err := opuslib.NewDecoderWithOutput(48000, 2)
	if err != nil {
		fmt.Println(err)
		return
	}

	rdr := &opusCodecReader{opusDecoder: opusdecoder}
	go rdr.Init()
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
		time.Sleep(time.Millisecond * 45)
		bpacket := [1024]byte{}
		packet := rtp.Packet{}
		for {

			n, _, err := rtpreader.Read(bpacket[:1024], nil)
			if err != nil {
				fmt.Println(err)
				if errors.Is(err, jitterbuffer.ErrPopWhileBuffering) {
					time.Sleep(4 * time.Millisecond)
					continue
				}
				rtp_read_stop <- err
				return
			}
			// packet, _, err := t.ReadRTP()
			err = packet.Unmarshal(bpacket[:n])
			if err != nil {
				log.Println(err)
				rtp_read_stop <- err
				return
			}

			_, err = rdr.Write(packet.Payload)
			if err != nil {
				log.Println(err)
				continue
			}
			// sb.Push(packet)

			// for sample := sb.Pop(); sample != nil; sample = sb.Pop() {
			// 	_, err := rdr.Write(sample.Data)

			// 	if err != nil {
			// 		fmt.Println(err)
			// 	}
			// }

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
	err = audio_c.StopAudio()
	if err != nil {
		fmt.Println("Error while trying to suspend audio", err)
	}
}

func handle_tracks(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
	fmt.Println("codec:", t.Codec())
	factory, err := jitterbuffer.NewInterceptor()
	if err != nil {
		log.Println(err)
		return
	}
	gen, err := factory.NewInterceptor("")
	if err != nil {
		log.Println(err)
		return
	}
	chain := interceptor.NewChain([]interceptor.Interceptor{gen})

	rtpreader := chain.BindRemoteStream(nil, interceptor.RTPReaderFunc(
		func(b []byte, _ interceptor.Attributes) (int, interceptor.Attributes, error) {
			t.SetReadDeadline(time.Now().Add(10 * time.Second))
			return t.Read(b)
		},
	))
	fmt.Print("recieving tracks")
	fmt.Print("Paylaod tpe", t.PayloadType())

	go func() {
		streamReader(t, rtpreader)

	}()
}
