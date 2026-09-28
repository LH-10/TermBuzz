package main

import (
	"github.com/LH-10/TermBuzz/STUN/stun"
)

func main() {
	var err error
	conn, err := stun.Listen("localhost")
	if err != nil {
		panic(err)
	}
	stun.Process(conn)
}
