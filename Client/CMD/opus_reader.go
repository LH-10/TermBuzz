package main

import (
	"fmt"

	opuslib "github.com/pion/opus"
)

type opusCodecReader struct {
	buf         []byte
	opusDecoder opuslib.Decoder
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
