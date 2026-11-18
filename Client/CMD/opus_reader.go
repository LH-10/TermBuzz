package main

import (
	"fmt"
	"sync/atomic"

	opuslib "github.com/pion/opus"
)

// keep the buffer size value as something to the power of 2 like 8 ,16 ,32
// this allows 'and' operation to reset the index in ring buffer to 0
const buffer_size = 0x80

type audio_buffer struct {
	buf []byte
}

func (ab *audio_buffer) Init() {
	ab.buf = make([]byte, 3850)
}
func (ab *audio_buffer) Write(p []byte) int {
	n := copy(ab.buf, p)
	return n
}

func (ab *audio_buffer) Read(p []byte) int {
	n := copy(p, ab.buf)
	return n
}

type opusCodecReader struct {
	opusDecoder opuslib.Decoder
	audio_buff  [buffer_size]audio_buffer
	write_index atomic.Uint32
	read_index  atomic.Uint32
	stopReading bool
}

func (ocr *opusCodecReader) Init() {
	for i := range ocr.audio_buff {
		ocr.audio_buff[i].Init()
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

	write_index := ocr.write_index.Load()
	read_index := ocr.read_index.Load()

	next_wri := (write_index + 1) & (buffer_size - 1) //resets indx to 0 , similar to indx % buffer_size
	if next_wri == read_index {
		return 0, fmt.Errorf("Buffer Full")
	}
	ocr.audio_buff[write_index].Write(out)
	ocr.write_index.Store(next_wri)
	return len(out), nil

}
func (ocr *opusCodecReader) HasData() bool {
	return ocr.write_index.Load() > 2
}

func (ocr *opusCodecReader) Read(p []byte) (n int, err error) {
	write_index := ocr.write_index.Load()
	read_index := ocr.read_index.Load()

	if read_index == write_index {
		if ocr.stopReading {
			return 0, fmt.Errorf("Reading Stopped")
		}
		return 0, nil
	}

	n = ocr.audio_buff[read_index].Read(p)
	ocr.read_index.Store((read_index + 1) & (buffer_size - 1)) //resets indx to 0
	return n, err

}
