package main

import (
	"fmt"
	"sync"

	opuslib "github.com/pion/opus"
)

const buffer_size = 0x10 //keep the buffer size value as something to the power of 2 like 8 ,16 ,32 this allows 'and' operation to reset the index in ring buffer to 0

type audio_buffer struct {
	buf []byte
	mx  sync.Mutex
}

func (ab *audio_buffer) Init() {
	ab.buf = make([]byte, 3850)
}
func (ab *audio_buffer) Write(p []byte) int {
	ab.mx.Lock()
	defer ab.mx.Unlock()
	n := copy(ab.buf, p)
	return n
}
func (ab *audio_buffer) LockedRead(p []byte) int {
	ab.mx.Lock()
	defer ab.mx.Unlock()
	n := copy(p, ab.buf)
	return n
}
func (ab *audio_buffer) Read(p []byte) int {

	n := copy(p, ab.buf)
	return n
}

type opusCodecReader struct {
	opusDecoder opuslib.Decoder
	abf         [buffer_size]audio_buffer
	indx        int
	readindx    int
	stopReading bool
}

func (ocr *opusCodecReader) Init() {
	for i := range ocr.abf {
		ocr.abf[i].Init()
	}
}
func (ocr *opusCodecReader) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	var out []byte = make([]byte, 3850) //2*1920+10
	_, _, err = ocr.opusDecoder.Decode(p, out)
	if err != nil {
		fmt.Println("here")
		fmt.Println("err:", err)
		return 0, err
	}

	ocr.abf[ocr.indx].Write(out)
	ocr.indx = (ocr.indx + 1) & (buffer_size - 1) //resets indx to 0 , similar to indx % buffer_size
	return len(out), nil

}
func (ocr *opusCodecReader) HasData() bool {
	return ocr.indx > 2
}

func (ocr *opusCodecReader) Read(p []byte) (n int, err error) {
	if ocr.readindx == ocr.indx {
		if ocr.stopReading {
			return 0, fmt.Errorf("Reading Stopped")
		}
		n = ocr.abf[ocr.readindx].LockedRead(p)
		ocr.readindx = (ocr.readindx + 1) & (buffer_size - 1)
		return
	}
	n = ocr.abf[ocr.readindx].Read(p)
	ocr.readindx = (ocr.readindx + 1) & (buffer_size - 1) //resets indx to 0 , similar to indx % buffer_size
	// if n == 0 {
	// 	return n, io.EOF
	// }
	return n, err

}
