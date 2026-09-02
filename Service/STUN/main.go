package main

import (
	"github.com/LH-10/TermBuzz/STUN/stun"
)

func main() {
	var err error
	conn, err := stun.Listen("127.0.0.1")
	if err != nil {
		panic(err)
	}
	stun.Process(conn)
}
