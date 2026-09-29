// SPDX-License-Identifier: GPL-2.0
//TermBuzz
// Copyright (C) 2026  Lalit Hinduja

//    This program is free software; you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation;

//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.

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
